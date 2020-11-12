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

// ResponseWriter is used to construct an HTTP response. When a Response is
// passed to the ResponseWriter, it will invoke the Dispatcher with the
// Response. An attempt to write to the ResponseWriter twice will
// cause a panic.
//
// A ResponseWriter may not be used after the Handler.ServeHTTP method has returned.
type ResponseWriter interface {
	// Write writes a safe response.
	Write(resp Response) Result

	// NoContent responds with a 204 No Content response.
	//
	// If the ResponseWriter has already been written to, then this method panics.
	NoContent() Result

	// WriteError writes an error response (400-599) according to the provided status code.
	//
	// If the ResponseWriter has already been written to, then this method panics.
	WriteError(code StatusCode) Result

	// Redirect responds with a redirect to the given url, using code as the status code.
	//
	// If the ResponseWriter has already been written to, then this method panics.
	Redirect(r *IncomingRequest, url string, code StatusCode) Result

	// Header returns the collection of headers that will be set
	// on the response. Headers must be set before writing a
	// response (e.g. Write, WriteTemplate).
	Header() Header

	// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
	// The provided cookie must have a valid Name, otherwise an error will be
	// returned.
	SetCookie(c *Cookie) error

	// SetCode allows setting a response status. If the response was already
	// written, trying to set the status code will have no effect. This method will
	// panic if an invalid status code is passed (i.e. not in the range 1XX-5XX).
	//
	// If SetCode was called before NoContent, Redirect or WriteError, the status
	// code set by the latter will be the actual response status.
	SetCode(code StatusCode)
}

type responseWriter struct {
	d  Dispatcher
	rw http.ResponseWriter

	// Having this field unexported is essential for security. Otherwise one can
	// easily overwrite the struct bypassing all our safety guarantees.
	code    StatusCode
	header  Header
	handler handler
	req     *IncomingRequest
	written bool
}

// NewResponseWriter creates a ResponseWriter from a safehttp.Dispatcher, an
// http.ResponseWriter and a reference to the current IncomingRequest being served.
// The IncomingRequest will only be used by the commit phase, which only runs when
// the ServeMux is used, and can be passed as nil in tests.
// TODO: remove the IncomingRequest parameter.
func NewResponseWriter(d Dispatcher, rw http.ResponseWriter, req *IncomingRequest) ResponseWriter {
	return newResponseWriter(d, rw, req)
}

func newResponseWriter(d Dispatcher, rw http.ResponseWriter, req *IncomingRequest) *responseWriter {
	header := newHeader(rw.Header())
	return &responseWriter{
		d:      d,
		rw:     rw,
		header: header,
		req:    req,
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

// Write dispatches the response to the Dispatcher. This will be written to the
// underlying http.ResponseWriter if the Dispatcher decides it's safe to do so.
//
// TODO: replace panics with proper error handling when writing the response fails.
func (w *responseWriter) Write(resp Response) Result {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.handler.commitPhase(w, resp)
	if w.written {
		return Result{}
	}

	w.markWritten()

	ct, err := w.d.ContentType(resp)
	if err != nil {
		panic(err)
	}
	w.rw.Header().Set("Content-Type", ct)

	if w.code != 0 {
		w.rw.WriteHeader(int(w.code))
	} else {
		w.rw.WriteHeader(int(StatusOK))
	}

	if err := w.d.Write(w.rw, resp); err != nil {
		panic(err)
	}
	return Result{}
}

// NoContent responds with a 204 No Content response.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (w *responseWriter) NoContent() Result {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.handler.commitPhase(w, NoContentResponse{})
	if w.written {
		return Result{}
	}
	w.markWritten()
	w.rw.WriteHeader(int(StatusNoContent))
	return Result{}
}

// WriteError writes an error response (400-599) according to the provided
// status code.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (w *responseWriter) WriteError(code StatusCode) Result {
	w.markWritten()
	resp := &ErrorResponse{Code: code}
	w.handler.errorPhase(w, resp)
	http.Error(w.rw, http.StatusText(int(resp.Code)), int(resp.Code))
	return Result{}
}

// Redirect responds with a redirect to the given url, using code as the status code.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (w *responseWriter) Redirect(r *IncomingRequest, url string, code StatusCode) Result {
	if code < 300 || code >= 400 {
		panic("wrong method called")
	}
	w.markWritten()
	http.Redirect(w.rw, r.req, url, int(code))
	return Result{}
}

// markWritten ensures that the ResponseWriter is only written to once by panicking
// if it is written more than once.
func (w *responseWriter) markWritten() {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.written = true
}

// Header returns the collection of headers that will be set
// on the response. Headers must be set before writing a
// response (e.g. Write, WriteTemplate).
func (w responseWriter) Header() Header {
	return w.header
}

// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name, otherwise an error will be
// returned.
func (w *responseWriter) SetCookie(c *Cookie) error {
	return w.header.addCookie(c)
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
func (w *responseWriter) SetCode(code StatusCode) {
	if w.written {
		return
	}
	if code < 100 || code >= 600 {
		panic("invalid status code")
	}
	w.code = code
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
