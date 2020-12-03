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

package xsrf_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf/xsrfhtml"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml/template"
)

func TestServeMuxInstallXSRF(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}
	it := xsrfhtml.Interceptor{SecretAppKey: "testSecretAppKey"}
	mb.Intercept(&it)

	handler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		fns := map[string]interface{}{
			"XSRFToken": func() string { return "WrongToken" },
		}
		t := template.Must(template.New("name").Funcs(fns).Parse(`<form><input type="hidden" name="token" value="{{XSRFToken}}">{{.}}</form>`))

		return safehttp.ExecuteTemplateWithFuncs(w, t, "Content", fns)
	})
	mb.Handle("/bar", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rr := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/bar", nil)

	mb.Mux().ServeHTTP(rr, req)

	if got, want := rr.Status(), safehttp.StatusOK; got != want {
		t.Errorf("rr.Status() got: %v want: %v", got, want)
	}

	if gotBody := b.String(); strings.Contains(gotBody, "WrongToken") {
		t.Errorf("response body: got %q, injection failed", gotBody)
	}

}
