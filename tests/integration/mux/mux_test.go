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

package safemux_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
	safetemplate "github.com/google/safehtml/template"
	"html/template"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type responseRecorder struct {
	headers http.Header
	writer  io.Writer
	status  safehttp.StatusCode
}

func newResponseRecorder(w io.Writer) *responseRecorder {
	return &responseRecorder{headers: http.Header{}, writer: w}
}

func (r *responseRecorder) Header() http.Header {
	return r.headers
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = safehttp.StatusCode(statusCode)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

func TestMuxDefaultDispatcher(t *testing.T) {
	tests := []struct {
		name        string
		mux         *safehttp.ServeMux
		wantStatus  safehttp.StatusCode
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
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
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
			wantStatus: safehttp.StatusInternalServerError,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Internal Server Error\n",
		},
		{
			name: "Safe HTML Template Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.WriteTemplate(safetemplate.Must(safetemplate.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "<h1>This is an actual heading, though.</h1>",
		},
		{
			name: "Unsafe Template Response",
			mux: func() *safehttp.ServeMux {
				mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")

				h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.WriteTemplate(template.Must(template.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
				})
				mux.Handle("/pizza", safehttp.MethodGet, h)
				return mux
			}(),
			wantStatus: safehttp.StatusInternalServerError,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Internal Server Error\n",
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
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"application/json; charset=utf-8"},
			},
			wantBody: "{\"field\":\"myField\"}\n",
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
			wantStatus: safehttp.StatusInternalServerError,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Internal Server Error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/pizza", nil)
			b := &strings.Builder{}
			rw := newResponseRecorder(b)

			tt.mux.ServeHTTP(rw, req)

			if rw.status != tt.wantStatus {
				t.Errorf("rw.status: got %v want %v", rw.status, tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.headers)); diff != "" {
				t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
			}

			if gotBody := b.String(); tt.wantBody != gotBody {
				t.Errorf("response body: got %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
