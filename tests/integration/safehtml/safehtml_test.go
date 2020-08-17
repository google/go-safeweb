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

package safehtml_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

type testDispatcher struct{}

func (testDispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (testDispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

type responseRecorder struct {
	header http.Header
	writer io.Writer
	status int
}

func newResponseRecorder(w io.Writer) *responseRecorder {
	return &responseRecorder{header: http.Header{}, writer: w, status: http.StatusOK}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

func TestHandleRequestWrite(t *testing.T) {
	mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
	mux.Handle("/", safehttp.MethodGet, safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
		return rw.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
	}))

	req := httptest.NewRequest("GET", "http://foo.com/", nil)

	b := &strings.Builder{}
	rw := newResponseRecorder(b)

	mux.ServeHTTP(rw, req)

	body := b.String()

	if want := "&lt;h1&gt;Escaped, so not really a heading&lt;/h1&gt;"; body != want {
		t.Errorf("body got: %q want: %q", body, want)
	}
}

func TestHandleRequestWriteTemplate(t *testing.T) {
	mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
	mux.Handle("/", safehttp.MethodGet, safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
		return rw.WriteTemplate(template.Must(template.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
	}))

	req := httptest.NewRequest("GET", "http://foo.com/", nil)

	b := &strings.Builder{}
	rw := newResponseRecorder(b)

	mux.ServeHTTP(rw, req)

	body := b.String()

	if want := "<h1>This is an actual heading, though.</h1>"; body != want {
		t.Errorf("body got: %q want: %q", body, want)
	}
}
