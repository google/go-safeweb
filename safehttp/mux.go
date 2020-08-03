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

// ServeMux is a safe HTTP request multiplexer that wraps http.ServeMux.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that most closely matches the
// URL.
//
// The multiplexer contains a list of allowed domains that will be matched
// against each incoming request. A different handler can be specified for every
// HTTP method supported at a registered pattern.
type ServeMux struct {
	mux        *http.ServeMux
	domains    map[string]bool
	dispatcher Dispatcher

	// Maps user-provided patterns to combined handlers which encapsulate
	// multiple handlers, each one associated with an HTTP method.
	handlers map[string]methodHandler
}

// NewServeMux allocates and returns a new ServeMux
// TODO(@mattias, @mara): make domains a variadic of string **literals**.
func NewServeMux(dispatcher Dispatcher, domains ...string) *ServeMux {
	d := map[string]bool{}
	for _, host := range domains {
		d[host] = true
	}
	return &ServeMux{
		mux:        http.NewServeMux(),
		domains:    d,
		dispatcher: dispatcher,
		handlers:   map[string]methodHandler{},
	}
}

// Handle registers a handler for the given pattern and method. If another
// handler is already registered for the same pattern and method, Handle panics.
func (m *ServeMux) Handle(pattern string, method string, h Handler) {
	ch, ok := m.handlers[pattern]
	if !ok {
		ch := methodHandler{
			h:       make(map[string]Handler),
			domains: m.domains,
			d:       m.dispatcher,
		}
		ch.h[method] = h

		m.handlers[pattern] = ch
		m.mux.Handle(pattern, ch)
		return
	}

	if _, ok := ch.h[method]; ok {
		panic("method already registered")
	}
	ch.h[method] = h
}

// ServeHTTP dispatches the request to the handler whose method matches the
// incoming request and whose pattern most closely matches the request URL.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// methodHandler is collection of handlers based on the request method.
type methodHandler struct {
	// Maps an HTTP method to its handler
	h       map[string]Handler
	domains map[string]bool
	d       Dispatcher
}

// ServeHTTP dispatches the request to the handler associated with
// the incoming request's method.
func (c methodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !c.domains[r.Host] {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	h, ok := c.h[r.Method]
	if !ok {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	h.ServeHTTP(newResponseWriter(c.d, w), newIncomingRequest(r))
}
