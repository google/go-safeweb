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

package csp_test

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	safetemplate "github.com/google/safehtml/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeMuxInstallCSP(t *testing.T) {
	mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")
	mux.Install(&csp.Interceptor{})

	handler := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		fns := map[string]interface{}{
			"CSPNonce": func() string { return "WrongNonce" },
		}
		t := safetemplate.Must(safetemplate.New("name").Funcs(fns).Parse(`<script nonce="{{CSPNonce}}" type="application/javascript">alert("script")</script><h1>{{.}}</h1>`))

		return w.WriteTemplate(t, "Content")
	})
	mux.Handle("/bar", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rr := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/bar", nil)

	mux.ServeHTTP(rr, req)

	if want := safehttp.StatusOK; rr.Status() != want {
		t.Errorf("rr.Status() got: %v want: %v", rr.Status(), want)
	}

	if strings.Contains(b.String(), "WrongNonce") {
		t.Errorf("response body: invalid csp nonce injected")
	}

}
