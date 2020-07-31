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
	"github.com/google/go-cmp/cmp"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeDispatcher struct{}

func (fakeDispatcher) Write(rw http.ResponseWriter, resp Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (fakeDispatcher) ExecuteTemplate(rw http.ResponseWriter, t Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

type fakeResponseWriter struct {
	header http.Header
	writer io.Writer
	status int
}

func newFakeResponseWriter(w io.Writer) *fakeResponseWriter {
	return &fakeResponseWriter{header: http.Header{}, writer: w, status: http.StatusOK}
}

func (r *fakeResponseWriter) Header() http.Header {
	return r.header
}

func (r *fakeResponseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *fakeResponseWriter) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

func TestValidHost(t *testing.T) {
	mux := NewServeMux(fakeDispatcher{}, "foo.com")

	h := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/", MethodGet, h)

	req := httptest.NewRequest("GET", "http://foo.com/", nil)
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.ServeHTTP(rw, req)

	if want := 200; rw.status != want {
		t.Errorf("rw.status: got %v want %v", rw.status, want)
	}

	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rw.header)); diff != "" {
		t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}
}

func TestInvalidHost(t *testing.T) {
	mux := NewServeMux(fakeDispatcher{}, "foo.com")

	h := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/", MethodGet, h)

	req := httptest.NewRequest("GET", "http://bar.com/", nil)
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.ServeHTTP(rw, req)

	if want := 404; rw.status != want {
		t.Errorf("rw.status: got %v want %v", rw.status, want)
	}

	wantedHeaders := map[string][]string{
		"Content-Type": {"text/plain; charset=utf-8"}, "X-Content-Type-Options": {"nosniff"},
	}

	if diff := cmp.Diff(wantedHeaders, map[string][]string(rw.header)); diff != "" {
		t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "Not Found\n"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}
}

func TestInvalidMethod(t *testing.T) {
	mux := NewServeMux(fakeDispatcher{}, "foo.com")

	h := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/", MethodGet, h)

	req := httptest.NewRequest("POST", "http://foo.com/", nil)
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.ServeHTTP(rw, req)

	if want := 405; rw.status != want {
		t.Errorf("rw.status: got %v want %v", rw.status, want)
	}

	wantedHeaders := map[string][]string{
		"Content-Type": {"text/plain; charset=utf-8"}, "X-Content-Type-Options": {"nosniff"},
	}

	if diff := cmp.Diff(wantedHeaders, map[string][]string(rw.header)); diff != "" {
		t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "Method Not Allowed\n"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}
}

func TestHandleTwoMethods(t *testing.T) {
	d := fakeDispatcher{}
	mux := NewServeMux(d, "foo.com")

	registeredGetHandler := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World! GET</h1>"))
	})
	registeredPostHandler := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World! POST</h1>"))
	})
	mux.Handle("/bar", MethodGet, registeredGetHandler)
	mux.Handle("/bar", MethodPost, registeredPostHandler)

	postReq := httptest.NewRequest("POST", "http://foo.com/bar", nil)
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.ServeHTTP(rw, postReq)

	if want := 200; rw.status != want {
		t.Errorf("rw.status: got %v want %v", rw.status, want)
	}

	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rw.header)); diff != "" {
		t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "&lt;h1&gt;Hello World! POST&lt;/h1&gt;"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}

	getReq := httptest.NewRequest("GET", "http://foo.com/bar", nil)
	b = &strings.Builder{}
	rw = newFakeResponseWriter(b)

	mux.ServeHTTP(rw, getReq)

	if want := 200; rw.status != want {
		t.Errorf("rw.status: got %v want %v", rw.status, want)
	}

	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rw.header)); diff != "" {
		t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "&lt;h1&gt;Hello World! GET&lt;/h1&gt;"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}
}
