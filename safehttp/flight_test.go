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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml"
)

type panickingInterceptor struct {
	before, commit, onError bool
}

func (p panickingInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	if p.before {
		panic("before")
	}
	return safehttp.NotWritten()
}

func (p panickingInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	if p.commit {
		panic("commit")
	}
}

func (p panickingInterceptor) OnError(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	if p.onError {
		panic("onError")
	}
}

func TestFlightInterceptorPanic(t *testing.T) {
	tests := []struct {
		desc        string
		interceptor panickingInterceptor
		wantPanic   bool
	}{
		{
			desc:        "panic in Before",
			interceptor: panickingInterceptor{before: true},
			wantPanic:   true,
		},
		{
			desc:        "panic in Commit",
			interceptor: panickingInterceptor{commit: true},
			wantPanic:   true,
		},
		{
			desc:        "panic in OnError, but handler finishes successfully, so it doesn't happen",
			interceptor: panickingInterceptor{onError: true},
			wantPanic:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			mb := &safehttp.ServeMuxConfig{}
			mb.Intercept(tc.interceptor)
			mb.Handle("/search", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				// IMPORTANT: We are setting the header here and expecting to be
				// cleared if a panic occurs.
				w.Header().Set("foo", "bar")
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			}))
			mux := mb.Mux()

			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/search", nil)
			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)

			defer func() {
				r := recover()
				if !tc.wantPanic {
					if r != nil {
						t.Fatalf("unexpected panic %v", r)
					}
					return
				}
				if r == nil {
					t.Fatal("expected panic")
				}
				// Good, the panic got propagated.
				if len(rw.Header()) > 0 {
					t.Errorf("ResponseWriter.Header() got %v, want empty", rw.Header())
				}
			}()
			mux.ServeHTTP(rw, req)
		})
	}
}

func TestFlightHandlerPanic(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}
	mb.Handle("/search", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		// IMPORTANT: We are setting the header here and expecting to be
		// cleared if a panic occurs.
		w.Header().Set("foo", "bar")
		panic("handler")
	}))
	mux := mb.Mux()

	req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/search", nil)
	b := &strings.Builder{}
	rw := safehttptest.NewTestResponseWriter(b)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic")
		}
		// Good, the panic got propagated.
		if len(rw.Header()) > 0 {
			t.Errorf("ResponseWriter.Header() got %v, want empty", rw.Header())
		}
	}()
	mux.ServeHTTP(rw, req)
}
