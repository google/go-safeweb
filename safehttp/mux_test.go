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
	mux.Handle("/", map[string]Handler{MethodGet: h})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "foo.com"
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.serveHTTP(rw, req)

	if got, want := b.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}
}

func TestInvalidHost(t *testing.T) {
	mux := NewServeMux(fakeDispatcher{}, "foo.com")

	h := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/", map[string]Handler{MethodGet: h})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "bar.com"
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.serveHTTP(rw, req)

	if got, want := b.String(), ""; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}

	if want := 403; rw.status != want {
		t.Errorf("rw.status: got %q want %q", rw.status, want)
	}
}

func TestInvalidMethod(t *testing.T) {
	mux := NewServeMux(fakeDispatcher{}, "foo.com")

	h := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/", map[string]Handler{MethodGet: h})

	req := httptest.NewRequest("POST", "/", nil)
	req.Host = "foo.com"
	b := &strings.Builder{}
	rw := newFakeResponseWriter(b)

	mux.serveHTTP(rw, req)

	if got, want := b.String(), ""; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}

	if want := 405; rw.status != want {
		t.Errorf("rw.status: got %q want %q", rw.status, want)
	}
}

func TestHandlerRegistered(t *testing.T) {
	d := fakeDispatcher{}
	mux := NewServeMux(d, "foo.com")

	registeredHandler := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/bar", map[string]Handler{MethodGet: registeredHandler})

	req := httptest.NewRequest("GET", "/bar", nil)
	req.Host = "foo.com"
	ir := newIncomingRequest(req)

	retrievedHandler, pattern := mux.Handler(&ir)

	if want := "/bar"; pattern != want {
		t.Errorf("mux.Handler(&ir) pattern got: %q want: %q", pattern, want)
	}

	b := &strings.Builder{}
	fakeRW := newFakeResponseWriter(b)

	rw := newResponseWriter(d, fakeRW)
	retrievedHandler.ServeHTTP(rw, &ir)

	if got, want := b.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf("response body: got %q want %q", got, want)
	}
}

func TestHandlerWrongMethod(t *testing.T) {
	d := fakeDispatcher{}
	mux := NewServeMux(d, "foo.com")

	registeredHandler := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/bar", map[string]Handler{MethodGet: registeredHandler})

	req := httptest.NewRequest("POST", "/bar", nil)
	req.Host = "foo.com"
	ir := newIncomingRequest(req)

	retrievedHandler, pattern := mux.Handler(&ir)

	if want := "/bar"; pattern != want {
		t.Errorf("mux.Handler(&ir) pattern got: %q want: %q", pattern, want)
	}

	b := &strings.Builder{}
	fakeRW := newFakeResponseWriter(b)

	rw := newResponseWriter(d, fakeRW)
	retrievedHandler.ServeHTTP(rw, &ir)

	if want := 404; fakeRW.status != want {
		t.Errorf("fakeRW.status: got %q want %q", fakeRW.status, want)
	}
}
