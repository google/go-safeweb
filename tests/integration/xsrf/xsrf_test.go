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
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml/template"
)

func TestServeMuxInstallXSRF(t *testing.T) {
	mb := &safehttp.ServeMuxBuilder{}
	it := xsrf.Interceptor{SecretAppKey: "testSecretAppKey"}
	mb.Install(&it)

	var token string
	var err error
	handler := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		fns := map[string]interface{}{
			"XSRFToken": func() string { return "WrongToken" },
		}
		// These are not used in the handler, just recorded to test whether the
		// handler can retrieve the XSRF token (e.g. for when the XSRF plugin is
		// installed, but auto-injection isn't yet supported and has to be done
		// manually by the handler).
		token, err = xsrf.Token(r)
		t := template.Must(template.New("name").Funcs(fns).Parse(`<form><input type="hidden" name="token" value="{{XSRFToken}}">{{.}}</form>`))

		return w.WriteTemplate(t, "Content")
	})
	mb.Handle("/bar", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rr := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/bar", nil)

	mb.Build().ServeHTTP(rr, req)

	if err != nil {
		t.Fatalf("xsrf.Token: got error %v", err)
	}

	if token == "" {
		t.Fatalf("csp.Nonce: got %q, want non-empty", token)
	}

	if got, want := rr.Status(), safehttp.StatusOK; got != want {
		t.Errorf("rr.Status() got: %v want: %v", got, want)
	}

	wantBody := `<form><input type="hidden" name="token" value="` +
		token + `">Content</form>`
	if gotBody := b.String(); gotBody != wantBody {
		t.Errorf("response body: got %q, want token %q", gotBody, wantBody)
	}

}
