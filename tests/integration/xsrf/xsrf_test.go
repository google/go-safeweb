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

package xsrf_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf/xsrfhtml"
	"github.com/google/safehtml/template"
)

func TestServeMuxInstallXSRF(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)
	it := xsrfhtml.Interceptor{SecretAppKey: "testSecretAppKey"}
	mb.Intercept(&it)
	mux := mb.Mux()

	handler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		fns := map[string]interface{}{
			"XSRFToken": func() string { return "WrongToken" },
		}
		t := template.Must(template.New("name").Funcs(fns).Parse(`<form><input type="hidden" name="token" value="{{XSRFToken}}">{{.}}</form>`))

		return safehttp.ExecuteTemplateWithFuncs(w, t, "Content", fns)
	})
	mux.Handle("/bar", safehttp.MethodGet, handler)

	rr := httptest.NewRecorder()

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/bar", nil)

	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, safehttp.StatusOK; got != int(want) {
		t.Errorf("rr.Code got: %v want: %v", got, int(want))
	}

	if gotBody := rr.Body.String(); strings.Contains(gotBody, "WrongToken") {
		t.Errorf("response body: got %q, injection failed", gotBody)
	}

}
