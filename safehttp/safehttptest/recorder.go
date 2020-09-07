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
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

type testDispatcher struct{}

func (testDispatcher) ContentType(resp safehttp.Response) (string, error) {
	switch resp.(type) {
	case safehtml.HTML, *template.Template:
		return "text/html; charset=utf-8", nil
	case safehttp.JSONResponse:
		return "application/json; charset=utf-8", nil
	default:
		return "", errors.New("not a safe response")
	}
}

func (testDispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (testDispatcher) WriteJSON(rw http.ResponseWriter, resp safehttp.JSONResponse) error {
	obj, err := json.Marshal(resp.Data)
	if err != nil {
		panic("invalid json")
	}
	_, err = rw.Write(obj)
	return err
}

func (testDispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

// ResponseRecorder encapsulates a safehttp.ResponseWriter that records
// mutations for later inspection in tests. The safehttp.ResponseWriter
// should be passed as part of the handler function in tests.
type ResponseRecorder struct {
	*safehttp.ResponseWriter
	rw *responseWriter
	b  *strings.Builder
}

// NewResponseRecorder creates a ResponseRecorder from the default testDispatcher.
func NewResponseRecorder() *ResponseRecorder {
	var b strings.Builder
	rw := newResponseWriter(&b)
	return &ResponseRecorder{
		rw:             rw,
		b:              &b,
		ResponseWriter: safehttp.NewResponseWriter(testDispatcher{}, rw, nil),
	}
}

// NewResponseRecorderFromDispatcher creates a ResponseRecorder from a
// provided safehttp.Dispatcher.
func NewResponseRecorderFromDispatcher(d safehttp.Dispatcher) *ResponseRecorder {
	var b strings.Builder
	rw := newResponseWriter(&b)
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

// responseWriter is an implementation of the http.ResponseWriter interface used
// for constructing an HTTP response.
type responseWriter struct {
	header http.Header
	writer io.Writer
	status safehttp.StatusCode
}

func newResponseWriter(w io.Writer) *responseWriter {
	return &responseWriter{
		header: http.Header{},
		writer: w,
		status: http.StatusOK,
	}
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.status = safehttp.StatusCode(statusCode)
}

func (r *responseWriter) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}
