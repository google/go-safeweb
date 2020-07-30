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
	"errors"
	"net/http"
)

// TODO: add the missing methods
const (
	// HTTP GET request
	MethodGet = "GET"
	// HTTP Post request
	MethodPost = "POST"
)

// ServeMux TODO
type ServeMux struct {
	mux        *http.ServeMux
	domains    map[string]bool
	dispatcher Dispatcher

	// Maps user-provided patterns to combined handlers which encapsulate
	// multiple handlers, each one associated with an HTTP method.
	handlers map[string]combinedHandler
}

// NewServeMux returns a safe multiplex containing the list of allowed domains
// and dispatcher provided by the user.
func NewServeMux(dispatcher Dispatcher, domains ...string) *ServeMux {
	d := map[string]bool{}
	for _, host := range domains {
		d[host] = true
	}
	return &ServeMux{mux: http.NewServeMux(), domains: d, dispatcher: dispatcher, handlers: map[string]combinedHandler{}}
}

// Handle registers a handler for the given pattern and method. The options
// are applied to the middleware. If another handler is already registered
// for the same pattern and method, Handle panics.
func (m *ServeMux) Handle(pattern string, h map[string]Handler) {
	ch := combinedHandler{h: h}
	m.handlers[pattern] = ch
	m.mux.Handle(pattern, ch.combine(m.domains, m.dispatcher))
}

// HandleFunc registers a handler for the given pattern and method. The options
// are applied to the middleware. If another handler is already registered
// for the same pattern and method, HandleFunc panics.
func (m *ServeMux) HandleFunc(pattern string, h map[string]HandleFunc) {
	if h == nil {
		return
	}
	newH := make(map[string]Handler)
	for key, val := range h {
		newH[key] = HandlerFunc(val)
	}
	m.Handle(pattern, newH)
}

// NotFound registers a handler which will be called when an unhandled
// path is visited. If another handler has already been registered for
// this purpose, NotFound panics.
func (*ServeMux) NotFound(h Handler) {
	panic("not implemented")
}

// NotFoundFunc registers a handler which will be called when an unhandled
// path is visited. If another handler has already been registered for
// this purpose, NotFound panics.
func (*ServeMux) NotFoundFunc(h HandleFunc) {
	panic("not implemented")
}

// Handler returns the handler to use for the incoming request and the pattern.
func (m *ServeMux) Handler(r *IncomingRequest) (Handler, string) {
	h, pattern := m.mux.Handler(r.req)

	if pattern == "" {
		// If the pattern is empty, no handler was registered and the handler is
		// an http.NotFoundHandler
		// TODO: replace http.NotFoundHandler to a safehttp.NotFoundHandler
		return safeHandler(h), pattern
	}

	// See if it is redirect or combined
	c, ok := m.handlers[pattern]
	if !ok {
		// We got a http.RedirectHandler from m.mux.Handler()
		return safeHandler(h), pattern
	}

	safeH, err := c.lookup(r.req.Method)
	if err != nil {
		// err is nil unless no HTTP method was registered for this pattern
		// TODO: replace http.NotFoundHandler to a safehttp.MethodNotAllowedHandler
		return safeHandler(http.NotFoundHandler()), pattern
	}

	return safeH, pattern
}

func (m *ServeMux) serveHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// combinedHandler is collection of handlers based on the request method.
type combinedHandler struct {
	// Maps an HTTP method to its handler
	h map[string]Handler
}

// lookup returns the handler associated with the HTTP method provided as an
// argument and a nil error, unless no handler was registered for the HTTP
// method.
func (c *combinedHandler) lookup(httpMethod string) (Handler, error) {
	h, ok := c.h[httpMethod]
	if !ok {
		return nil, errors.New("method not registered")
	}
	return h, nil
}

// combine creates a combined handler to be registered with http.ServeMux.Handle
// which calls the correct safe handler based on the request method.
func (c combinedHandler) combine(domains map[string]bool, d Dispatcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !domains[r.Host] {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h, ok := c.h[r.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		rw := newResponseWriter(d, w)
		ir := newIncomingRequest(r)
		h.ServeHTTP(rw, &ir)
	})
}
