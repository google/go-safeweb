// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safehttptest

import (
	"net/http"
	"net/http/httptest"

	"github.com/google/go-safeweb/safehttp"
)

// FakeResponseWriter creates a fake safehttp.ResponseWriter implementation.
//
// It performs no error checking nor runs interceptors.
type FakeResponseWriter struct {
	// The Dispatcher implementation
	Dispatcher *FakeDispatcher

	// ResponseWriter is only used for calls to Dispatcher. Calls to AddCookie()
	// do not affect it.
	ResponseWriter http.ResponseWriter

	// Cookies coming from AddCookie() calls. Use the ResponseWriter to see what
	// cookies have been set by the Dispatcher.
	Cookies []*safehttp.Cookie

	// Response headers.
	Headers safehttp.Header
}

// FakeDispatcher provides a minimal implementation of the Dispatcher to be used for testing Interceptors.
type FakeDispatcher struct {
	Written    safehttp.Response
	Dispatcher safehttp.Dispatcher
}

// Write records the written responses and calls Dispatcher.Write.
// If Dispatcher is nil, the default one is used.
func (d *FakeDispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	d.Written = resp
	if d.Dispatcher != nil {
		return d.Dispatcher.Write(rw, resp)
	}
	return safehttp.DefaultDispatcher{}.Write(rw, resp)
}

// Error writes just the status code.
func (d *FakeDispatcher) Error(rw http.ResponseWriter, resp safehttp.ErrorResponse) error {
	rw.WriteHeader(int(resp.Code()))
	return nil
}

// NewFakeResponseWriter creates a new safehttp.ResponseWriter implementation
// and a httptest.ResponseRecorder, for testing purposes only.
func NewFakeResponseWriter() (*FakeResponseWriter, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return &FakeResponseWriter{
		Dispatcher:     &FakeDispatcher{},
		ResponseWriter: recorder,
		Headers:        safehttp.NewHeader(recorder.HeaderMap),
	}, recorder
}

var _ safehttp.ResponseWriter = (*FakeResponseWriter)(nil)

// Header returns the Header.
func (frw *FakeResponseWriter) Header() safehttp.Header {
	return frw.Headers
}

// AddCookie appends the given cookie to the Cookies field.
func (frw *FakeResponseWriter) AddCookie(c *safehttp.Cookie) error {
	if len(c.Name()) == 0 {
		panic("empty cookie name")
	}

	frw.Cookies = append(frw.Cookies, c)
	return nil
}

// Write forwards the response to Dispatcher.Write.
func (frw *FakeResponseWriter) Write(resp safehttp.Response) safehttp.Result {
	if err := frw.Dispatcher.Write(frw.ResponseWriter, resp); err != nil {
		panic(err)
	}
	return safehttp.Result{}
}

// NoContent writes just the NoContent status code.
func (frw *FakeResponseWriter) NoContent() safehttp.Result {
	frw.ResponseWriter.WriteHeader(int(safehttp.StatusNoContent))
	return safehttp.Result{}
}

// WriteError forwards the error response to Dispatcher.WriteError.
func (frw *FakeResponseWriter) WriteError(resp safehttp.ErrorResponse) safehttp.Result {
	if err := frw.Dispatcher.Error(frw.ResponseWriter, resp); err != nil {
		panic(err)
	}
	return safehttp.Result{}
}
