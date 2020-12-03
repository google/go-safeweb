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

package hsts_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp/safehttptest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/hsts"
	"github.com/google/safehtml"
)

func TestHSTSServeMuxInstall(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}

	mb.Intercept(hsts.Default())
	handler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mb.Handle("/asdf", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rw := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(http.MethodGet, "https://foo.com/asdf", nil)

	mb.Mux().ServeHTTP(rw, req)

	if want := safehttp.StatusOK; rw.Status() != want {
		t.Errorf("rw.Status() got: %v want: %v", rw.Status(), want)
	}

	wantHeaders := map[string][]string{
		"Content-Type":              {"text/html; charset=utf-8"},
		"Strict-Transport-Security": {"max-age=63072000; includeSubDomains"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf("b.String() got: %v want: %v", got, want)
	}
}

func TestHSTSOnErrors(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}

	mb.Intercept(hsts.Default())
	handler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.WriteError(safehttp.StatusNotFound)
	})
	mb.Handle("/asdf", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rw := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(http.MethodGet, "https://foo.com/asdf", nil)

	mb.Mux().ServeHTTP(rw, req)

	if want := safehttp.StatusNotFound; rw.Status() != want {
		t.Errorf("rw.Status() got: %v want: %v", rw.Status(), want)
	}
	if got, want := rw.Header()["Strict-Transport-Security"], []string{"max-age=63072000; includeSubDomains"}; !cmp.Equal(got, want) {
		t.Errorf("rw.Header()[\"Strict-Transport-Security\"] = %q, want %q", got, want)
	}
}
