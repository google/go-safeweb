// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safehttp_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/go-safeweb/safehttp"
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

func (panickingInterceptor) Match(safehttp.InterceptorConfig) bool {
	return false
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
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			mb := safehttp.NewServeMuxConfig(nil)
			mb.Intercept(tc.interceptor)
			mux := mb.Mux()

			mux.Handle("/search", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				// IMPORTANT: We are setting the header here and expecting to be
				// cleared if a panic occurs.
				w.Header().Set("foo", "bar")
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			}))

			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/search", nil)
			rw := httptest.NewRecorder()

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
	mb := safehttp.NewServeMuxConfig(nil)
	mux := mb.Mux()

	mux.Handle("/search", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		// IMPORTANT: We are setting the header here and expecting to be
		// cleared if a panic occurs.
		w.Header().Set("foo", "bar")
		panic("handler")
	}))

	req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/search", nil)
	rw := httptest.NewRecorder()

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

func TestFlightDoubleWritePanics(t *testing.T) {
	writeFuncs := map[string]func(safehttp.ResponseWriter, *safehttp.IncomingRequest) safehttp.Result{
		"Write": func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
			return w.Write(safehtml.HTMLEscaped("Hello"))
		},
		"WriteError": func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
			return w.WriteError(safehttp.StatusPreconditionFailed)
		},
	}

	for firstWriteName, firstWrite := range writeFuncs {
		for secondWriteName, secondWrite := range writeFuncs {
			t.Run(fmt.Sprintf("%s->%s", firstWriteName, secondWriteName), func(t *testing.T) {
				mb := safehttp.NewServeMuxConfig(nil)
				mux := mb.Mux()
				mux.Handle("/search", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					firstWrite(w, r)
					secondWrite(w, r) // this should panic
					t.Fatal("should never reach this point")
					return safehttp.Result{}
				}))

				req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/search", nil)
				rw := httptest.NewRecorder()
				defer func() {
					if r := recover(); r == nil {
						t.Fatalf("expected panic")
					}
					// Good, the panic got propagated.
					// Note: we are not testing the response headers here, as the first write might have already succeeded.
				}()
				mux.ServeHTTP(rw, req)
			})

		}
	}

}
