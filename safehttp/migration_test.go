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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

func TestRegisteredHandler(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)

	mb.Handle("/abc", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("Welcome!"))
	}))
	mb.Handle("/abc", safehttp.MethodPost, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		f, err := r.PostForm()
		if err != nil {
			return w.WriteError(safehttp.StatusBadRequest)
		}
		animal := f.String("animal", "")
		if animal == "" {
			return w.WriteError(safehttp.StatusBadRequest)
		}
		return w.Write(safehtml.HTMLEscaped(fmt.Sprintf("Added %s.", animal)))
	}))
	mb.Handle("/def", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("Bye!"))
	}))

	safeMux := mb.Mux()

	mux := http.NewServeMux()
	mux.Handle("/abc", safehttp.RegisteredHandler(safeMux, "/abc"))
	mux.Handle("/def", safehttp.RegisteredHandler(safeMux, "/def"))

	var tests = []struct {
		name       string
		req        *http.Request
		wantStatus safehttp.StatusCode
		wantBody   string
	}{
		{
			name:       "Valid GET Request",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://foo.com/abc", nil),
			wantStatus: safehttp.StatusOK,
			wantBody:   "Welcome!",
		},
		{
			name: "Valid POST Request",
			req: func() *http.Request {
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/abc", strings.NewReader("animal=cat"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			wantStatus: safehttp.StatusOK,
			wantBody:   "Added cat.",
		},
		{
			name:       "Different handler",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://foo.com/def", nil),
			wantStatus: safehttp.StatusOK,
			wantBody:   "Bye!",
		},
		{
			name:       "Invalid Method",
			req:        httptest.NewRequest(safehttp.MethodHead, "http://foo.com/abc", nil),
			wantStatus: safehttp.StatusMethodNotAllowed,
			wantBody:   "Method Not Allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := httptest.NewRecorder()

			mux.ServeHTTP(rw, tt.req)

			if rw.Code != int(tt.wantStatus) {
				t.Errorf("rw.Code: got %v want %v", rw.Code, tt.wantStatus)
			}

			if got := rw.Body.String(); got != tt.wantBody {
				t.Errorf("response body: got %q want %q", got, tt.wantBody)
			}
		})
	}
}

func TestRegisteredHandler_StrictPatterns(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)

	mb.Handle("/foo/", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("Homepage!"))
	}))
	safeMux := mb.Mux()

	if safehttp.RegisteredHandler(safeMux, "/foo/") == nil {
		t.Error(`RegisteredHandler(_, "/foo/") got nil, want non-nil`)
	}
	if safehttp.RegisteredHandler(safeMux, "/foo/subpath") != nil {
		t.Error(`RegisteredHandler(_, "/foo/subpath") got non-nil, want nil`)
	}
}
