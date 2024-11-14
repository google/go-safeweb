// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mux_test

import (
	"html/template"
	"math"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
	safetemplate "github.com/google/safehtml/template"
)

func TestMuxDefaultDispatcher(t *testing.T) {
	tests := []struct {
		name        string
		handler     safehttp.Handler
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "Safe HTML Response",
			handler: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			}),
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name: "Safe HTML Template Response",
			handler: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return safehttp.ExecuteTemplate(w, safetemplate.
					Must(safetemplate.New("name").
						Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
			}),
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "<h1>This is an actual heading, though.</h1>",
		},
		{
			name: "Valid JSON Response",
			handler: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				data := struct {
					Field string `json:"field"`
				}{Field: "myField"}
				return safehttp.WriteJSON(w, data)
			}),
			wantHeaders: map[string][]string{
				"Content-Type": {"application/json; charset=utf-8"},
			},
			wantBody: ")]}',\n{\"field\":\"myField\"}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := safehttp.NewServeMuxConfig(nil)
			mux := mb.Mux()

			mux.Handle("/pizza", safehttp.MethodGet, tt.handler)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/pizza", nil)

			mux.ServeHTTP(rw, req)

			if wantStatus := safehttp.StatusOK; rw.Code != int(wantStatus) {
				t.Errorf("rw.Code: got %v want %v", rw.Code, wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header mismatch (-want +got):\n%s", diff)
			}

			if gotBody := rw.Body.String(); tt.wantBody != gotBody {
				t.Errorf("response body: got %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestMuxDefaultDispatcherUnsafeResponses(t *testing.T) {
	tests := []struct {
		name    string
		handler safehttp.Handler
	}{
		{
			name: "Unsafe HTML Response",
			handler: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write("<h1>Hello World!</h1>")
			}),
		},
		{
			name: "Unsafe Template Response",
			handler: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return safehttp.ExecuteTemplate(w, template.
					Must(template.New("name").
						Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
			}),
		},
		{
			name: "Invalid JSON Response",
			handler: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return safehttp.WriteJSON(w, math.Inf(1))
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Unskip these test cases and combine them with the test
			// cases from the previous test into a single table test after
			// error-handling in the ResponseWriter has been fixed.
			t.Skip()

			mb := safehttp.NewServeMuxConfig(nil)
			mux := mb.Mux()

			mux.Handle("/pizza", safehttp.MethodGet, tt.handler)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/pizza", nil)

			mux.ServeHTTP(rw, req)

			if wantStatus := safehttp.StatusInternalServerError; rw.Code != int(wantStatus) {
				t.Errorf("rw.Code: got %v want %v", rw.Code, wantStatus)
			}

			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header(): mismatch (-want +got):\n%s", diff)
			}

			if wantBody, gotBody := "Internal Server Error\n", rw.Body.String(); wantBody != gotBody {
				t.Errorf("response body: got %v, want %v", gotBody, wantBody)
			}
		})
	}
}
