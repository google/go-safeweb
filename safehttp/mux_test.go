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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

type testDispatcher struct{}

func (testDispatcher) Write(c ResponseWriterContainer, resp Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		rw := c.Release(StatusOK, "text/html; charset=utf-8")
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (testDispatcher) ExecuteTemplate(c ResponseWriterContainer, t Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		rw := c.Release(StatusOK, "text/html; charset=utf-8")
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

func TestMuxOneHandlerOneRequest(t *testing.T) {
	var test = []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid Request",
			req:        httptest.NewRequest("GET", "http://foo.com/", nil),
			wantStatus: 200,
			wantHeader: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name:       "Invalid Host",
			req:        httptest.NewRequest("GET", "http://bar.com/", nil),
			wantStatus: 404,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Not Found\n",
		},
		{
			name:       "Invalid Method",
			req:        httptest.NewRequest("POST", "http://foo.com/", nil),
			wantStatus: 405,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Method Not Allowed\n",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mux := NewServeMux(testDispatcher{}, "foo.com")

			h := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mux.Handle("/", MethodGet, h)

			b := &strings.Builder{}
			rw := newResponseRecorder(b)

			mux.ServeHTTP(rw, tt.req)

			if rw.status != tt.wantStatus {
				t.Errorf("rw.status: got %v want %v", rw.status, tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeader, map[string][]string(rw.header)); diff != "" {
				t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
			}

			if got := b.String(); got != tt.wantBody {
				t.Errorf("response body: got %q want %q", got, tt.wantBody)
			}
		})
	}
}

func TestMuxServeTwoHandlers(t *testing.T) {
	var tests = []struct {
		name        string
		req         *http.Request
		hf          HandlerFunc
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "GET Handler",
			req:  httptest.NewRequest("GET", "http://foo.com/bar", nil),
			hf: HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World! GET</h1>"))
			}),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World! GET&lt;/h1&gt;",
		},
		{
			name: "POST Handler",
			req:  httptest.NewRequest("POST", "http://foo.com/bar", nil),
			hf: HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World! POST</h1>"))
			}),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World! POST&lt;/h1&gt;",
		},
	}

	d := testDispatcher{}
	mux := NewServeMux(d, "foo.com")
	mux.Handle("/bar", MethodGet, tests[0].hf)
	mux.Handle("/bar", MethodPost, tests[1].hf)

	for _, test := range tests {
		b := &strings.Builder{}
		rw := newResponseRecorder(b)
		mux.ServeHTTP(rw, test.req)
		if want := test.wantStatus; rw.status != want {
			t.Errorf("rw.status: got %v want %v", rw.status, want)
		}

		if diff := cmp.Diff(test.wantHeaders, map[string][]string(rw.header)); diff != "" {
			t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
		}

		if got, want := b.String(), test.wantBody; got != want {
			t.Errorf("response body: got %q want %q", got, want)
		}
	}
}

func TestMuxHandleSameMethodTwice(t *testing.T) {
	d := testDispatcher{}
	mux := NewServeMux(d, "foo.com")

	registeredHandler := HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/bar", MethodGet, registeredHandler)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf(`mux.Handle("/bar", MethodGet, registeredHandler) expected panic`)
		}
	}()

	mux.Handle("/bar", MethodGet, registeredHandler)
}
