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

package mux_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml"
	safetemplate "github.com/google/safehtml/template"
	"html/template"
	"math"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMuxDefaultDispatcher(t *testing.T) {
	tests := []struct {
		name        string
		mux         *safehttp.ServeMux
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "Safe HTML Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name: "Safe HTML Template Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.WriteTemplate(safetemplate.
						Must(safetemplate.New("name").
							Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "<h1>This is an actual heading, though.</h1>",
		},
		{
			name: "Valid JSON Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					data := struct {
						Field string `json:"field"`
					}{Field: "myField"}
					return w.WriteJSON(data)
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
			wantHeaders: map[string][]string{
				"Content-Type": {"application/json; charset=utf-8"},
			},
			wantBody: ")]}',\n{\"field\":\"myField\"}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/pizza", nil)
			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)

			tt.mux.ServeHTTP(rw, req)

			if wantStatus := safehttp.StatusOK; rw.Status() != wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
			}

			if gotBody := b.String(); tt.wantBody != gotBody {
				t.Errorf("response body: got %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestMuxDefaultDispatcherUnsafeResponses(t *testing.T) {
	tests := []struct {
		name string
		mux  *safehttp.ServeMux
	}{
		{
			name: "Unsafe HTML Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.Write("<h1>Hello World!</h1>")
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
		},
		{
			name: "Unsafe Template Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.WriteTemplate(template.
						Must(template.New("name").
							Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
		},
		{
			name: "Invalid JSON Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.WriteJSON(math.Inf(1))
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Unskip these test cases and combine them with the test
			// cases from the previous test into a single table test after
			// error-handling in the ResponseWriter has been fixed.
			t.Skip()
			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/pizza", nil)
			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)

			tt.mux.ServeHTTP(rw, req)

			if wantStatus := safehttp.StatusInternalServerError; rw.Status() != wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), wantStatus)
			}

			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header(): mismatch (-want +got):\n%s", diff)
			}

			if wantBody, gotBody := "Internal Server Error\n", b.String(); wantBody != gotBody {
				t.Errorf("response body: got %v, want %v", gotBody, wantBody)
			}
		})
	}
}
