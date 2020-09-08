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
	header           Header
	appliedInterceps []AppliedInterceptor
	req              *IncomingRequest
	written          bool
}

// NewResponseWriter creates a ResponseWriter from a safehttp.Dispatcher, an
// http.ResponseWriter and a list of interceptors associated with a ServeMux.
func NewResponseWriter(d Dispatcher, rw http.ResponseWriter, req *IncomingRequest, interceps []AppliedInterceptor) *ResponseWriter {
	header := newHeader(rw.Header())
	return &ResponseWriter{
		d:                d,
		rw:               rw,
		header:           header,
		appliedInterceps: interceps,
		req:              req,
	}
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

// commitPhase calls the Commit phases of all the interceptors. This stage will
// run before a response is written to the ResponseWriter. If a response is
// written to the ResponseWriter in a Commit phase then the Commit phases of the
// remaining interceptors won't execute.
//
// TODO: BIG WARNING, if ResponseWriter.Write and ResponseWriter.WriteTemplate
// are called in Commit then this will recurse. CommitResponseWriter was an
// attempt to prevent this by not giving access to Write and WriteTemplate in
// the Commit phase.
func (w *ResponseWriter) commitPhase(resp Response) {
	for i := len(w.appliedInterceps) - 1; i >= 0; i-- {
		ai := w.appliedInterceps[i]
		ai.it.Commit(w, w.req, resp, ai.cfg)
		if w.written {
			return
		}
	}
}

// Write TODO
func (w *ResponseWriter) Write(resp Response) Result {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.commitPhase(resp)
	if w.written {
		return Result{}
	}
	if err := w.d.Write(w.rw, resp); err != nil {
		panic("error")
	}
	w.markWritten()
	return Result{}
}

// WriteTemplate TODO
func (w *ResponseWriter) WriteTemplate(t Template, data interface{}) Result {
	if w.written {
		panic("ResponseWriter was already written to")
	}
	w.commitPhase(TemplateResponse{Template: &t, Data: &data})
	if w.written {
		return Result{}
	}
	if err := w.d.ExecuteTemplate(w.rw, t, data); err != nil {
		panic("error")
	}
	w.markWritten()
	return Result{}
}

// NoContent responds with a 204 No Content response.
func (w *ResponseWriter) NoContent() Result {
	w.markWritten()
	w.rw.WriteHeader(int(StatusNoContent))
	return Result{}
}

// WriteError writes an error response (400-599) according to the provided status
// code. Any headers previously set will be removed.
func (w *ResponseWriter) WriteError(code StatusCode) Result {
	w.markWritten()
	http.Error(w.rw, http.StatusText(int(code)), int(code))
	return Result{}
}

// Redirect responds with a redirect to a given url, using code as the status code.
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
