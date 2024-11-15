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

package staticheaders_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/staticheaders"
	"github.com/google/safehtml"
)

func TestServeMuxInstallStaticHeaders(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)
	mb.Intercept(staticheaders.Interceptor{})
	mux := mb.Mux()

	handler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/asdf", safehttp.MethodGet, handler)

	rw := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "https://foo.com/asdf", nil)

	mux.ServeHTTP(rw, req)

	if want := safehttp.StatusOK; rw.Code != int(want) {
		t.Errorf("rw.Code got: %v want: %v", rw.Code, want)
	}

	wantHeaders := map[string][]string{
		"Content-Type":           {"text/html; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
		"X-Xss-Protection":       {"0"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}

	if got, want := rw.Body.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf("response body got: %v want: %v", got, want)
	}
}

func TestStaticHeadersOnError(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)
	mb.Intercept(staticheaders.Interceptor{})
	mux := mb.Mux()

	handler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.WriteError(safehttp.StatusNotFound)
	})
	mux.Handle("/asdf", safehttp.MethodGet, handler)

	rw := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "https://foo.com/asdf", nil)

	mux.ServeHTTP(rw, req)

	if want := safehttp.StatusNotFound; rw.Code != int(want) {
		t.Errorf("rw.Status() got: %v want: %v", rw.Code, want)
	}

	wantHeaders := map[string][]string{
		"X-Content-Type-Options": {"nosniff"},
		"X-Xss-Protection":       {"0"},
	}
	for h, want := range wantHeaders {
		if got := rw.Header()[h]; !cmp.Equal(got, want) {
			t.Errorf("rw.Header()[%q] got %q, want %q", h, got, want)
		}
	}
}
