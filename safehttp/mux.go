// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safehttp

import (
	"net/http"
)

// TODO: add the missing methods
const (
	// HTTP GET request
	MethodGet = "GET"
	// HTTP Post request
	MethodPost = "POST"
)

// ServeMux is an HTTP request multiplexer. It matches the URL of each incoming
// request against a list of registered patterns and calls the handler for
// the pattern that most closely matches the URL.
//
// When creating the multiplexer, the user needs to specify a list of allowed
// domains. The server will only serve requests target to those domains and
// otherwise will reply with HTTP 404 Not Found.
// TODO(@mihalimara22, @mattiasgrenfeldt): add a link to docs/ explaining
// why this is done.
//
// Patterns names are fixed, rooted paths, like "/favicon.ico", or rooted
// subtrees like "/images/" (note the trailing slash). Longer patterns take
// precedence over shorter ones, so that if there are handlers registered for
// both "/images/" and "/images/thumbnails/", the latter handler will be called
// for paths beginning "/images/thumbnails/" and the former will receive
// requests for any other paths in the "/images/" subtree.
//
// Note that since a pattern ending in a slash names a rooted subtree, the
// pattern "/" matches all paths not matched by other registered patterns,
// not just the URL with Path == "/".
//
// If a subtree has been registered and a request is received naming the subtree
// root without its trailing slash, ServeMux redirects that request to
// the subtree root (adding the trailing slash). This behavior can be overridden
// with a separate registration for the path without the trailing slash. For
// example, registering "/images/" causes ServeMux to redirect a request for
// "/images" to "/images/", unless "/images" has been registered separately.
//
// Patterns may optionally begin with a host name, restricting matches to URLs
// on that host only. This host name must be in the list of allowed domains passed
// when creating the ServeMux. Host-specific patterns take precedence over general
// patterns, so that a handler might register for the two patterns "/codesearch"
// and "codesearch.google.com/" without also taking over requests for
// "http://www.google.com/".
//
// ServeMux also takes care of sanitizing the URL request path and the Host
// header, stripping the port number and redirecting any request containing . or
// .. elements or repeated slashes to an equivalent, cleaner URL.
//
// Multiple handlers can be registered for a single pattern, as long as they
// handle different HTTP methods.
type ServeMux struct {
	mux     *http.ServeMux
	domains map[string]bool
	disp    Dispatcher

	// Maps patterns to handlers supporting multiple HTTP methods.
	handlers map[string]methodHandler

	// Maps interceptor key to interceptor.
	interceps map[string]Interceptor
}

// NewServeMux allocates and returns a new ServeMux
func NewServeMux(d Dispatcher, domains ...string) *ServeMux {
	// TODO(@mattiasgrenfeldt, @mihalimara22): make domains a variadic of string **literals**.
	dm := map[string]bool{}
	for _, host := range domains {
		dm[host] = true
	}
	return &ServeMux{
		mux:       http.NewServeMux(),
		domains:   dm,
		disp:      d,
		handlers:  map[string]methodHandler{},
		interceps: map[string]Interceptor{},
	}
}

// Handle registers a handler for the given pattern and method. If another
// handler is already registered for the same pattern and method, Handle panics.
func (m *ServeMux) Handle(pattern string, method string, h Handler) {
	mh, ok := m.handlers[pattern]
	if !ok {
		mh := methodHandler{
			handlers:     map[string]Handler{method: h},
			domains:      m.domains,
			disp:         m.disp,
			muxInterceps: m.interceps,
		}

		m.handlers[pattern] = mh
		m.mux.Handle(pattern, mh)
		return
	}

	if _, ok := mh.handlers[method]; ok {
		panic("method already registered")
	}
	mh.handlers[method] = h
}

// Install installs an Interceptor. Interceptor keys need to be unique. If an
// Interceptor with the same key has already been installed, Install panics.
//
// TODO(@empijei, @grenfeldt, @kele, @mihalimara22): Right now you could install
// the same interceptor twice with different keys, we need to figure out how
// exactly we want to avoid that and how we define key uniqueness.
func (m *ServeMux) Install(key string, i Interceptor) {
	if _, exists := m.interceps[key]; exists {
		panic("interceptor with same key already installed")
	}
	m.interceps[key] = i
}

// ServeHTTP dispatches the request to the handler whose method matches the
// incoming request and whose pattern most closely matches the request URL.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// methodHandler is a collection of handlers based on the request method.
type methodHandler struct {
	// Maps an HTTP method to its handler
	handlers     map[string]Handler
	domains      map[string]bool
	disp         Dispatcher
	muxInterceps map[string]Interceptor
}

// ServeHTTP dispatches the request to the handler associated with
// the incoming request's method after calling the Before function of all
// ServeMux interceptors the handler is registered on.
func (m methodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(m.disp, w, m.muxInterceps)
	ir := NewIncomingRequest(r)

	if !m.domains[r.Host] {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	h, ok := m.handlers[r.Method]
	if !ok {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			rw.ServerError(StatusInternalServerError)
		}
	}()

	for _, intercep := range m.muxInterceps {
		if res := intercep.Before(rw, ir); res.written {
			return
		}
	}

	h.ServeHTTP(rw, ir)
}
