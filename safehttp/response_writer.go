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

// ResponseWriter TODO
type ResponseWriter struct {
	d  Dispatcher
	rw http.ResponseWriter

	// Having this field unexported is essential for security. Otherwise one can
	// easily overwrite the struct bypassing all our safety guarantees.
	header       Header
	muxInterceps map[string]Interceptor
	written      bool
}

// NewResponseWriter creates a ResponseWriter from a safehttp.Dispatcher, an
// http.ResponseWriter and a list of interceptors associated with a ServeMux.
func NewResponseWriter(d Dispatcher, rw http.ResponseWriter, muxInterceps map[string]Interceptor) *ResponseWriter {
	header := newHeader(rw.Header())
	return &ResponseWriter{
		d:            d,
		rw:           rw,
		header:       header,
		muxInterceps: muxInterceps,
	}
}

// Interceptor returns the interceptor associated with the given key.
func (w *ResponseWriter) Interceptor(key string) Interceptor {
	mp, ok := w.muxInterceps[key]
	if !ok {
		return nil
	}
	return mp
}

// Result TODO
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

// Write TODO
func (w *ResponseWriter) Write(resp Response) Result {
	w.markWritten()
	if err := w.d.Write(w.rw, resp); err != nil {
		panic("error")
	}
	return Result{}
}

// WriteTemplate TODO
func (w *ResponseWriter) WriteTemplate(t Template, data interface{}) Result {
	w.markWritten()
	if err := w.d.ExecuteTemplate(w.rw, t, data); err != nil {
		panic("error")
	}
	return Result{}
}

// NoContent responds with a 204 No Content response.
func (w *ResponseWriter) NoContent() Result {
	w.markWritten()
	w.rw.WriteHeader(int(StatusNoContent))
	return Result{}
}

// ClientError TODO
func (w *ResponseWriter) ClientError(code StatusCode) Result {
	w.markWritten()
	if code < 400 || code >= 500 {
		// TODO(@mihalimara22): Replace this when we decide how to handle this case
		panic("wrong method called")
	}
	http.Error(w.rw, http.StatusText(int(code)), int(code))
	return Result{}
}

// ServerError TODO
func (w *ResponseWriter) ServerError(code StatusCode) Result {
	w.markWritten()
	if code < 500 || code >= 600 {
		// TODO(@mattiasgrenfeldt, @mihalimara22, @kele, @empijei): Decide how it should
		// be communicated to the user of the framework that they've called the wrong
		// method.
		panic("wrong method called")
	}
	http.Error(w.rw, http.StatusText(int(code)), int(code))
	return Result{}
}

// Redirect responds with a redirect to a given url, using code as the status code.
func (w *ResponseWriter) Redirect(r *IncomingRequest, url string, code StatusCode) Result {
	w.markWritten()
	if code < 300 || code >= 400 {
		panic("wrong method called")
	}
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
// The provided cookie must have a valid Name. Otherwise an error will be
// returned.
func (w *ResponseWriter) SetCookie(c *Cookie) error {
	return w.header.addCookie(c)
}

// Dispatcher TODO
type Dispatcher interface {
	Write(rw http.ResponseWriter, resp Response) error
	ExecuteTemplate(rw http.ResponseWriter, t Template, data interface{}) error
}
