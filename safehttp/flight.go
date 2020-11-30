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

	cfg HandlerConfig

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
// of the HandlerConfig needed by the flight. For the existing tests that use it,
// it's fine, but they should be migrated.
//
// TODO(kele): remove this once we have a better option to provide a
// ResponseWriter implementation for interceptor testing.
func DeprecatedNewResponseWriter(rw http.ResponseWriter, dispatcher Dispatcher) ResponseWriter {
	return &flight{
		cfg:    HandlerConfig{Dispatcher: dispatcher},
		rw:     rw,
		header: newHeader(rw.Header()),
	}
}

// HandlerConfig is the safe HTTP handler configuration, including the
// dispatcher and interceptors.
type HandlerConfig struct {
	Handler      Handler
	Dispatcher   Dispatcher
	Interceptors []ConfiguredInterceptor
}

func processRequest(cfg HandlerConfig, rw http.ResponseWriter, req *http.Request) {
	f := &flight{
		cfg:    cfg,
		rw:     rw,
		header: newHeader(rw.Header()),
		req:    NewIncomingRequest(req),
	}

	// The net/http package handles all panics. In the early days of the
	// framework we were handling them ourselves and running interceptors after
	// a panic happened, but this adds lots of complexity to the codebase and
	// still isn't perfect (e.g. what if OnError panics?). Instead, we just make
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

	ct, err := f.cfg.Dispatcher.ContentType(resp)
	if err != nil {
		panic(err)
	}
	f.rw.Header().Set("Content-Type", ct)

	if f.code == 0 {
		f.code = StatusOK
	}
	f.rw.WriteHeader(int(f.code))

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
func (f *flight) WriteError(code StatusCode) Result {
	// TODO: accept custom error responses that need to go through the dispatcher.
	if f.written {
		panic("ResponseWriter was already written to")
	}
	f.written = true
	resp := &ErrorResponse{Code: code}
	f.errorPhase(resp)
	http.Error(f.rw, http.StatusText(int(resp.Code)), int(resp.Code))
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

// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name, otherwise an error will be
// returned.
//
// TODO: should this be named AddCookie?
func (f *flight) SetCookie(c *Cookie) error {
	return f.header.addCookie(c)
}

// SetCode allows setting a response status. If the response was already
// written, trying to set the status code will have no effect. This method will
// panic if an invalid status code is passed (i.e. not in the range 1XX-5XX).
//
// If SetCode was called before NoContent, Redirect or WriteError, the status
// code set by the latter will be the actual response status.
//
// TODO(empijei@, kele@, maramihali@): decide what should be done if the
// code passed is either 3XX (redirect) or 4XX-5XX (client/server error).
func (f *flight) SetCode(code StatusCode) {
	if f.written {
		return
	}
	if code < 100 || code >= 600 {
		panic("invalid status code")
	}
	f.code = code
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

// errorPhase calls the OnError phases of all the interceptors associated with
// a handler. This stage runs before an error response is written to the
// ResponseWriter.
func (f *flight) errorPhase(resp Response) {
	for i := len(f.cfg.Interceptors) - 1; i >= 0; i-- {
		f.cfg.Interceptors[i].OnError(f, f.req, resp)
	}
}

// Dispatcher is responsible for writing a response received from the
// ResponseWriter to the underlying http.ResponseWriter.
//
// The implementation of a custom Dispatcher should be thoroughly reviewed by
// the security team to avoid introducing vulnerabilities.
type Dispatcher interface {
	// Content-Type returns the Content-Type of the provided response if it is
	// of a safe type, supported by the Dispatcher, and should return an error
	// otherwise.
	//
	// Sending a response to the http.ResponseWriter without properly setting
	// CT is error-prone and could introduce vulnerabilities. Therefore, this
	// method should be used to set the Content-Type header before calling
	// Dispatcher.Write. Writing should not proceed if ContentType returns an
	// error.
	ContentType(resp Response) (string, error)

	// Write writes a Response to the underlying http.ResponseWriter.
	//
	// It should return an error if the writing operation fails or if the
	// provided Response should not be written to the http.ResponseWriter
	// because it's unsafe.
	Write(rw http.ResponseWriter, resp Response) error
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
