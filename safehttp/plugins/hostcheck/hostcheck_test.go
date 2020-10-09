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

package hostcheck_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/hostcheck"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml"
)

func TestInterceptor(t *testing.T) {
	var test = []struct {
		name       string
		req        *http.Request
		wantStatus safehttp.StatusCode
	}{
		{
			name:       "Valid Host",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://foo.com/", nil),
			wantStatus: safehttp.StatusOK,
		},
		{
			name:       "Invalid Host",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://bar.com/", nil),
			wantStatus: safehttp.StatusNotFound,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mb := &safehttp.ServeMuxConfig{}
			mb.Intercept(hostcheck.New("foo.com"))

			h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mb.Handle("/", safehttp.MethodGet, h)

			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)
			mux := mb.Mux()
			mux.ServeHTTP(rw, tt.req)

			if rw.Status() != tt.wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), tt.wantStatus)
			}
		})
	}
}
