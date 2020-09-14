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

package safehttptest

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/go-safeweb/safehttp"
)

// ResponseRecorder encapsulates a safehttp.ResponseWriter that records
// mutations for later inspection in tests. The safehttp.ResponseWriter
// should be passed as part of the handler function in tests.
type ResponseRecorder struct {
	*safehttp.ResponseWriter
	rw *TestResponseWriter
	b  *strings.Builder
}

// NewResponseRecorder creates a ResponseRecorder from the safehttp.DefaultDispatcher.
func NewResponseRecorder() *ResponseRecorder {
	var b strings.Builder
	rw := NewTestResponseWriter(&b)
	return &ResponseRecorder{
		rw:             rw,
		b:              &b,
		ResponseWriter: safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, rw, nil),
	}
}

// NewResponseRecorderFromDispatcher creates a ResponseRecorder from a
// provided safehttp.Dispatcher.
func NewResponseRecorderFromDispatcher(d safehttp.Dispatcher) *ResponseRecorder {
	var b strings.Builder
	rw := NewTestResponseWriter(&b)
	return &ResponseRecorder{
		rw:             rw,
		b:              &b,
		ResponseWriter: safehttp.NewResponseWriter(d, rw, nil),
	}
}

// Header returns the recorded response headers.
func (r *ResponseRecorder) Header() http.Header {
	return r.rw.Header()
}

// Status returns the recorded response status code.
func (r *ResponseRecorder) Status() safehttp.StatusCode {
	return r.rw.status
}

// Body returns the recorded response body.
func (r *ResponseRecorder) Body() string {
	return r.b.String()
}

// TestResponseWriter is an implementation of the http.ResponseWriter interface
// used for constructing an HTTP response for testing purposes.
type TestResponseWriter struct {
	header http.Header
	writer io.Writer
	status safehttp.StatusCode
}

// NewTestResponseWriter creates a new TestResponseWriter.
func NewTestResponseWriter(w io.Writer) *TestResponseWriter {
	return &TestResponseWriter{
		header: http.Header{},
		writer: w,
		status: safehttp.StatusOK,
	}
}

// Status returns the response status.
func (r *TestResponseWriter) Status() safehttp.StatusCode {
	return r.status
}

// Header implements http.ResponseWriter. It returns the response headers that
// could have been mutated within a handler.
func (r *TestResponseWriter) Header() http.Header {
	return r.header
}

// WriteHeader implements http.ResponseWriter.
func (r *TestResponseWriter) WriteHeader(statusCode int) {
	r.status = safehttp.StatusCode(statusCode)
}

// Write implements http.ResponseWriter.
func (r *TestResponseWriter) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}
