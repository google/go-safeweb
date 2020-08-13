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

package safehttp_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

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
			wantHeader: map[string][]string{},
			wantBody:   "&lt;h1&gt;Hello World!&lt;/h1&gt;",
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
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")

			h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mux.Handle("/", safehttp.MethodGet, h)

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
		hf          safehttp.Handler
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "GET Handler",
			req:  httptest.NewRequest("GET", "http://foo.com/bar", nil),
			hf: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World! GET</h1>"))
			}),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "&lt;h1&gt;Hello World! GET&lt;/h1&gt;",
		},
		{
			name: "POST Handler",
			req:  httptest.NewRequest("POST", "http://foo.com/bar", nil),
			hf: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World! POST</h1>"))
			}),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "&lt;h1&gt;Hello World! POST&lt;/h1&gt;",
		},
	}

	d := testDispatcher{}
	mux := safehttp.NewServeMux(d, "foo.com")
	mux.Handle("/bar", safehttp.MethodGet, tests[0].hf)
	mux.Handle("/bar", safehttp.MethodPost, tests[1].hf)

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
	mux := safehttp.NewServeMux(d, "foo.com")

	registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/bar", safehttp.MethodGet, registeredHandler)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf(`mux.Handle("/bar", MethodGet, registeredHandler) expected panic`)
		}
	}()

	mux.Handle("/bar", safehttp.MethodGet, registeredHandler)
}
