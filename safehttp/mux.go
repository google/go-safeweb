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
	"fmt"
	"net/http"
)

// The HTTP request methods defined by RFC.
const (
	MethodConnect = "CONNECT" // RFC 7231, 4.3.6
	MethodDelete  = "DELETE"  // RFC 7231, 4.3.5
	MethodGet     = "GET"     // RFC 7231, 4.3.1
	MethodHead    = "HEAD"    // RFC 7231, 4.3.2
	MethodOptions = "OPTIONS" // RFC 7231, 4.3.7
	MethodPatch   = "PATCH"   // RFC 5789
	MethodPost    = "POST"    // RFC 7231, 4.3.3
	MethodPut     = "PUT"     // RFC 7231, 4.3.4
	MethodTrace   = "TRACE"   // RFC 7231, 4.3.8
)

// ServeMux is an HTTP request multiplexer. It matches the URL of each incoming
// request against a list of registered patterns and calls the handler for
// the pattern that most closely matches the URL.
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
// on that host only.  Host-specific patterns take precedence over general
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
	mux *http.ServeMux
}

func registerHandlers(mux *http.ServeMux, handlers map[string]map[string]handler) {
	for pattern, handlersPerMethod := range handlers {
		pattern := pattern
		handlersPerMethod := handlersPerMethod
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			h, ok := handlersPerMethod[r.Method]
			if !ok {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// ServeHTTP dispatches the request to the handler whose method matches the
// incoming request and whose pattern most closely matches the request URL.
//
//  For each incoming request:
//  - [Before Phase] Interceptor.Before methods are called for every installed
//    interceptor, until an interceptor writes to a ResponseWriter (including
//    errors) or panics,
//  - the handler is called after a [Before Phase] if no writes or panics occured,
//  - the handler triggers the [Commit Phase] by writing to the ResponseWriter,
//  - [Commit Phase] Interceptor.Commit methods run for every interceptor whose
//    Before method was called,
//  - [Dispatcher Phase] after the [Commit Phase], the Dispatcher's appropriate
//    write method is called; the Dispatcher is responsible for determining whether
//    the response is indeed safe and writing it,
//  - if the handler attempts to write more than once, it is treated as an
//    unrecoverable error; the request processing ends abrubptly with a panic and
//    nothing else happens (note: this will change as soon as [After Phase] is
//    introduced)
//
// Interceptors should NOT rely on the order they're run.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// ServeMuxConfig is a builder for ServeMux.
type ServeMuxConfig struct {
	dispatcher   Dispatcher
	handlers     []handlerRegistration
	interceptors []Interceptor
}

type handlerRegistration struct {
	pattern string
	method  string
	handler Handler
	cfgs    []InterceptorConfig
}

// Handle registers a handler for the given pattern and method. If a handler is
// registered twice for the same pattern and method, Build will panic.
//
// InterceptorConfigs can be passed in order to modify the behavior of the
// interceptors on a registered handler. Passing an InterceptorConfig whose
// corresponding Interceptor was not installed will produce no effect. If
// multiple configurations are passed for the same Interceptor, only the first
// one will take effect.
func (s *ServeMuxConfig) Handle(pattern string, method string, h Handler, cfgs ...InterceptorConfig) {
	s.handlers = append(s.handlers, handlerRegistration{
		pattern: pattern,
		method:  method,
		handler: h,
		cfgs:    cfgs,
	})
}

// Intercept installs an Interceptor.
//
// Interceptors order is undetermined and should not be relied on.
func (s *ServeMuxConfig) Intercept(i Interceptor) {
	s.interceptors = append(s.interceptors, i)
}

// Mux returns the ServeMux with a copy of the current configuration.
func (s *ServeMuxConfig) Mux() *ServeMux {
	dispatcher := s.dispatcher
	if dispatcher == nil {
		dispatcher = DefaultDispatcher{}
	}
	// pattern -> method -> handler
	handlers := map[string]map[string]handler{}
	for _, hr := range s.handlers {
		if handlers[hr.pattern] == nil {
			handlers[hr.pattern] = map[string]handler{}
		}
		if _, ok := handlers[hr.pattern][hr.method]; ok {
			panic(fmt.Sprintf("double registration of (pattern = %q, method = %q)", hr.pattern, hr.method))
		}
		handlers[hr.pattern][hr.method] = handler{
			handler:   hr.handler,
			interceps: configureInterceptors(s.interceptors, hr.cfgs),
			disp:      dispatcher,
		}
	}
	m := http.NewServeMux()
	registerHandlers(m, handlers)
	return &ServeMux{mux: m}
}

func configureInterceptors(interceptors []Interceptor, cfgs []InterceptorConfig) []ConfiguredInterceptor {
	var its []ConfiguredInterceptor
	for _, it := range interceptors {
		var cfg InterceptorConfig
		for _, c := range cfgs {
			if c.Match(it) {
				cfg = c
				break
			}
		}
		its = append(its, ConfiguredInterceptor{interceptor: it, config: cfg})
	}
	return its
}

// Clone creates a copy of the current config.
// This can be used to create several instances of Mux that share the same set of
// plugins and some common handlers.
func (s *ServeMuxConfig) Clone() *ServeMuxConfig {
	c := &ServeMuxConfig{
		dispatcher:   s.dispatcher,
		handlers:     make([]handlerRegistration, len(s.handlers)),
		interceptors: make([]Interceptor, len(s.interceptors)),
	}
	copy(c.handlers, s.handlers)
	copy(c.interceptors, s.interceptors)
	return c
}

// handler encapsulates a handler and its corresponding
// interceptors.
type handler struct {
	handler   Handler
	interceps []ConfiguredInterceptor
	disp      Dispatcher
}

// ServeHTTP calls the Before method of all the interceptors and then calls the
// underlying handler. Any panics that occur during the Before phase, during handler
// execution or during the Commit phase are recovered here and a 500 Internal Server
// Error response is sent.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ir := NewIncomingRequest(r)
	rw := newResponseWriter(h.disp, w, ir)
	rw.handler = h

	// The `net/http` package recovers handler panics, but we cannot rely on that behavior here.
	// The reason is, we might need to run After/Commit stages of the interceptors before we
	// respond with a 500 Internal Server Error.
	defer func() {
		if r := recover(); r != nil {
			rw.WriteError(StatusInternalServerError)
		}
	}()

	for _, it := range h.interceps {
		it.Before(rw, ir)
		if rw.written {
			return
		}
	}

	h.handler.ServeHTTP(rw, ir)
	if !rw.written {
		rw.NoContent()
	}
}

// commitPhase calls the Commit phases of all the interceptors. This stage will
// run before a response is written to the ResponseWriter. If a response is
// written to the ResponseWriter in a Commit phase then the Commit phases of the
// remaining interceptors won't execute.
//
// TODO: BIG WARNING, if ResponseWriter.Write and ResponseWriter.WriteTemplate
// are called in Commit then this will recurse. CommitResponseWriter was an
// attempt to prevent this by not giving access to Write and WriteTemplate in
// the Commit phase.
func (h handler) commitPhase(w *responseWriter, resp Response) {
	for i := len(h.interceps) - 1; i >= 0; i-- {
		it := h.interceps[i]
		it.Commit(w, w.req, resp)
		if w.written {
			return
		}
	}
}

// errorPhase calls the OnError phases of all the interceptors associated with
// a handler. This stage runs before an error response is written to the
// ResponseWriter.
//
// TODO: BIG WARNING, if an interceptor attempts to write to the ResponseWriter
// in the OnError phase will result in an irrecoverable error.
func (h handler) errorPhase(w *responseWriter, resp Response) {
	for i := len(h.interceps) - 1; i >= 0; i-- {
		it := h.interceps[i]
		it.OnError(w, w.req, resp)
	}
}
