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

// A single request "flight".
type flight struct {
	rw  http.ResponseWriter
	req *IncomingRequest

	cfg handlerConfig

	code   StatusCode
	header Header

	written bool
}

// DeprecatedNewResponseWriter creates a ResponseWriter implementation that has
// been historically used for testing. DO NOT USE this function, it's going
// to be removed soon.
//
// The main problem with this is the implementation of the returned
// ResponseWriter is incomplete, i.e. lacks the Handler and Interceptors fields
// of the handlerConfig needed by the flight. For the existing tests that use it,
// it's fine, but they should be migrated.
//
// TODO(kele): remove this once we have a better option to provide a
// ResponseWriter implementation for interceptor testing.
func DeprecatedNewResponseWriter(rw http.ResponseWriter, dispatcher Dispatcher) ResponseWriter {
	if dispatcher == nil {
		dispatcher = DefaultDispatcher{}
	}
	return &flight{
		cfg:    handlerConfig{Dispatcher: dispatcher},
		rw:     rw,
		header: newHeader(rw.Header()),
	}
}

// handlerConfig is the safe HTTP handler configuration, including the
// dispatcher and interceptors.
type handlerConfig struct {
	Handler      Handler
	Dispatcher   Dispatcher
	Interceptors []configuredInterceptor
}

func processRequest(cfg handlerConfig, rw http.ResponseWriter, req *http.Request) {
	f := &flight{
		cfg:    cfg,
		rw:     rw,
		header: newHeader(rw.Header()),
		req:    NewIncomingRequest(req),
	}

	// The net/http package handles all panics. In the early days of the
	// framework we were handling them ourselves and running interceptors after
	// a panic happened, but this adds lots of complexity to the codebase and
	// still isn't perfect (e.g. what if Commit panics?). Instead, we just make
	// sure to clear all the headers and cookies.
	defer func() {
		if r := recover(); r != nil {
			// Clear all headers.
			for h := range f.rw.Header() {
				delete(f.rw.Header(), h)
			}
			panic(r)
		}
	}()

	for _, it := range f.cfg.Interceptors {
		it.Before(f, f.req)
		if f.written {
			return
		}
	}
	f.cfg.Handler.ServeHTTP(f, f.req)
	if !f.written {
		f.NoContent()
	}
}

// Write dispatches the response to the Dispatcher. This will be written to the
// underlying http.ResponseWriter if the Dispatcher decides it's safe to do so.
func (f *flight) Write(resp Response) Result {
	if f.written {
		panic("ResponseWriter was already written to")
	}
	f.written = true
	f.commitPhase(resp)

	if err := f.cfg.Dispatcher.Write(f.rw, resp); err != nil {
		panic(err)
	}
	return Result{}
}

// NoContent responds with a 204 No Content response.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (f *flight) NoContent() Result {
	// TODO: should NoContent call Write under the hood? Should the dispatcher
	// handle this too?
	if f.written {
		panic("ResponseWriter was already written to")
	}
	f.written = true
	f.commitPhase(NoContentResponse{})
	f.rw.WriteHeader(int(StatusNoContent))
	return Result{}
}

// WriteError writes an error response (400-599) according to the provided
// status code.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (f *flight) WriteError(resp ErrorResponse) Result {
	if f.written {
		panic("ResponseWriter was already written to")
	}
	f.written = true
	f.commitPhase(resp)
	if err := f.cfg.Dispatcher.Error(f.rw, resp); err != nil {
		panic(err)
	}
	return Result{}
}

// Redirect responds with a redirect to the given url, using code as the status code.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (f *flight) Redirect(r *IncomingRequest, url string, code StatusCode) Result {
	if code < 300 || code >= 400 {
		panic(fmt.Sprintf("wrong method called: redirect with status %d", code))
	}
	if f.written {
		panic("ResponseWriter was already written to")
	}
	f.written = true
	http.Redirect(f.rw, r.req, url, int(code))
	return Result{}
}

// Header returns the collection of headers that will be set on the response.
// Headers must be set before writing a response.
func (f *flight) Header() Header {
	return f.header
}

// AddCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name, otherwise an error will be
// returned.
func (f *flight) AddCookie(c *Cookie) error {
	return f.header.addCookie(c)
}

// commitPhase calls the Commit phases of all the interceptors. This stage will
// run before a response is written to the ResponseWriter. If a response is
// written to the ResponseWriter in a Commit phase then the Commit phases of the
// remaining interceptors won'f execute.
func (f *flight) commitPhase(resp Response) {
	for i := len(f.cfg.Interceptors) - 1; i >= 0; i-- {
		f.cfg.Interceptors[i].Commit(f, f.req, resp)
	}
}

// Result is the result of writing an HTTP response.
//
// Use ResponseWriter methods to obtain it.
type Result struct{}

// NotWritten returns a Result which indicates that nothing has been written yet. It
// can be used in all functions that return a Result, such as in the ServeHTTP method
// of a Handler or in the Before method of an Interceptor. When returned, NotWritten
// indicates that the writing of the response should take place later. When this
// is returned by the Before method in Interceptors the next Interceptor in line
// is run. When this is returned by a Handler, a 204 No Content response is written.
func NotWritten() Result {
	return Result{}
}
