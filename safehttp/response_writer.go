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
type ResponseWriter struct {
	d  Dispatcher
	rw http.ResponseWriter

	// Having this field unexported is essential for security. Otherwise one can
	// easily overwrite the struct bypassing all our safety guarantees.
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
func NewResponseWriter(d Dispatcher, rw http.ResponseWriter, req *IncomingRequest) *ResponseWriter {
	header := newHeader(rw.Header())
	return &ResponseWriter{
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

func (w *ResponseWriter) Written() bool {
	if w.written {
		return true
	}
	return false
}

// Write dispatches the response to the Dispatcher, setting the Content-Type and
// response status if the provided response is a safe response. The
// Dispatcher will then write the response to the underlying Response Writer.
//
// TODO: replace panics with proper error handling when getting the response
// Content-Type or writing the response fails.
func (w *ResponseWriter) Write(resp Response) Result {
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
	w.rw.WriteHeader(int(StatusOK))
	if err := w.d.Write(w.rw, resp); err != nil {
		panic(err)
	}
	return Result{}
}

// WriteJSON encapsulates data into a JSON response and dispatches it to the
// Dispatcher, to be serialised and written to the ResponseWriter. It will also
// set the Content-Type and response status will be set.
//
// TODO: replace panics with proper error handling when getting the response
// Content-Type or writing the response fails.
func (w *ResponseWriter) WriteJSON(data interface{}) Result {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.handler.commitPhase(w, JSONResponse{Data: data})
	if w.written {
		return Result{}
	}
	w.markWritten()
	resp := JSONResponse{Data: data}
	ct, err := w.d.ContentType(resp)
	if err != nil {
		panic(err)
	}
	w.rw.Header().Set("Content-Type", ct)
	w.rw.WriteHeader(int(StatusOK))
	if err := w.d.WriteJSON(w.rw, resp); err != nil {
		panic(err)
	}
	return Result{}
}

// WriteTemplate dispatches a parsed template and a data object to the
// Dispatcher to be executed and written to the underlying Response Writer, in
// case the template is a safe HTML template. If it is a safe HTML Template, the
// Content-Type  and response status will also be set.
//
// TODO: replace panics with proper error handling when getting the response
// Content-Type or writing the response fails.
func (w *ResponseWriter) WriteTemplate(t Template, data interface{}) Result {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	resp := TemplateResponse{Template: &t, Data: &data, FuncMap: map[string]interface{}{}}
	w.handler.commitPhase(w, resp)
	if w.written {
		return Result{}
	}
	w.markWritten()
	ct, err := w.d.ContentType(t)
	if err != nil {
		panic(err)
	}
	w.rw.Header().Set("Content-Type", ct)
	w.rw.WriteHeader(int(StatusOK))
	if err := w.d.ExecuteTemplate(w.rw, resp); err != nil {
		panic(err)
	}
	return Result{}
}

// NoContent responds with a 204 No Content response.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (w *ResponseWriter) NoContent() Result {
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

// WriteError writes an error response (400-599) according to the provided status
// code.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (w *ResponseWriter) WriteError(code StatusCode) Result {
	w.markWritten()
	http.Error(w.rw, http.StatusText(int(code)), int(code))
	return Result{}
}

// Redirect responds with a redirect to the given url, using code as the status code.
//
// If the ResponseWriter has already been written to, then this method will panic.
func (w *ResponseWriter) Redirect(r *IncomingRequest, url string, code StatusCode) Result {
	if code < 300 || code >= 400 {
		panic("wrong method called")
	}
	w.markWritten()
	http.Redirect(w.rw, r.req, url, int(code))
	return Result{}
}

// markWritten ensures that the ResponseWriter is only written to once by panicking
// if it is written more than once.
func (w *ResponseWriter) markWritten() {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.written = true
}

// Header returns the collection of headers that will be set
// on the response. Headers must be set before writing a
// response (e.g. Write, WriteTemplate).
func (w ResponseWriter) Header() Header {
	return w.header
}

// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name, otherwise an error will be
// returned.
func (w *ResponseWriter) SetCookie(c *Cookie) error {
	return w.header.addCookie(c)
}

// Dispatcher is responsible for writing a response received from the
// ResponseWriter to the underlying http.ResponseWriter.
//
// The implementation of a custom Dispatcher should be thoroughly reviewed by
// the security team to avoid introducing vulnerabilities.
type Dispatcher interface {
	// Write writes a Response to the http.ResponseWriter.
	//
	// It should return an error if the writing operation fails or if the
	// provided Response should not be written to the http.ResponseWriter
	Write(rw http.ResponseWriter, resp Response) error

	// WriteJSON serialises and writes a JSON object to the http.ResponseWriter.
	//
	// This should return an error if the provided response is not a safe
	// JSONResponse  or if writing the object to the http.ResponseWriter fails.
	WriteJSON(rw http.ResponseWriter, resp JSONResponse) error

	// ExecuteTemplate applies the response parsed template to its data object
	// and writes the output to the http.ResponseWriter. If the response funcMap
	// is non-nil, the Dispatcher will try to update the names to functions
	// mappings of the template.
	//
	// This should return an error if the template contained by the
	// TemplateResponse is not a safe template or if an error occurs executing
	// or writing the template.
	ExecuteTemplate(rw http.ResponseWriter, resp TemplateResponse) error

	// Content-Type returns the Content-Type of the provided response if it is
	// of a type supported by the Dispatcher and should return an error
	// otherwise.
	//
	// Sending a response to the http.ResponseWriter without properly setting
	// CT is error-prone and could introduce vulnerabilities. Therefore, this
	// method should be used to set the Content-Type header before calling any
	// of the writing methods in the Dispatcher. Writing should not proceed
	// if ContentType returns an error.
	ContentType(resp Response) (string, error)
}
