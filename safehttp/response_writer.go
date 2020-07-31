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

	// Having this field unexported is essential for
	// security. Otherwise one can easily overwrite
	// the struct bypassing all our safety guarantees.
	header Header
}

func newResponseWriter(d Dispatcher, rw http.ResponseWriter) ResponseWriter {
	header := newHeader(rw.Header())
	return ResponseWriter{d: d, rw: rw, header: header}
}

// Result TODO
type Result struct{}

// Write TODO
func (w *ResponseWriter) Write(resp Response) Result {
	if err := w.d.Write(w.rw, resp); err != nil {
		panic("error")
	}
	return Result{}
}

// WriteTemplate TODO
func (w *ResponseWriter) WriteTemplate(t Template, data interface{}) Result {
	if err := w.d.ExecuteTemplate(w.rw, t, data); err != nil {
		panic("error")
	}
	return Result{}
}

// ServerError TODO
func (w *ResponseWriter) ServerError(code StatusCode, resp Response) Result {
	return Result{}
}

// Redirect responds with a redirect to the given url.
// The type of redirect is specified via the status code.
func (w *ResponseWriter) Redirect(r *IncomingRequest, url string, code StatusCode) Result {
	http.Redirect(w.rw, r.req, url, int(code))
	return Result{}
}

// Header returns the collection of headers that will be set
// on the response. Headers must be set before writing a
// response (e.g. Write, WriteTemplate).
func (w ResponseWriter) Header() Header {
	return w.header
}

// Dispatcher TODO
type Dispatcher interface {
	Write(rw http.ResponseWriter, resp Response) error
	ExecuteTemplate(rw http.ResponseWriter, t Template, data interface{}) error
}
