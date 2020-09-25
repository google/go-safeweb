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
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	safetemplate "github.com/google/safehtml/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeMuxInstallXSRF(t *testing.T) {
	mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{}, "foo.com")
	it := xsrf.Interceptor{SecretAppKey: "testSecretAppKey"}
	mux.Install(&it)

	var token string
	var err error
	handler := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		fns := map[string]interface{}{
			"XSRFToken": func() string { return "WrongToken" },
		}
		token, err = xsrf.Token(r)
		t := safetemplate.Must(safetemplate.New("name").Funcs(fns).Parse(`<form><input type="hidden" name="token" value="{{XSRFToken}}">{{.}}</form>`))

		return w.WriteTemplate(t, "Content")
	})
	mux.Handle("/bar", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rr := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/bar", nil)

	mux.ServeHTTP(rr, req)

	if err != nil {
		t.Fatalf("xsrf.Token: got error %v", err)
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
