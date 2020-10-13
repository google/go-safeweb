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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml/template"
)

func TestServeMuxInstallCSP(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}
	it := csp.Default("")
	mb.Intercept(&it)

	var nonce string
	var err error
	handler := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		fns := map[string]interface{}{
			"CSPNonce": func() string { return "WrongNonce" },
		}
		// These are not used in the handler, just recorded to test whether the
		// handler can retrieve the CSP nonce (e.g. for when the CSP plugin is
		// installed, but auto-injection isn't yet supported and has to be done
		// manually by the handler).
		nonce, err = csp.Nonce(r.Context())
		t := template.Must(template.New("name").Funcs(fns).Parse(`<script nonce="{{CSPNonce}}" type="application/javascript">alert("script")</script><h1>{{.}}</h1>`))

		return w.Write(safehttp.NewTemplateResponse(t, "Content", fns))
	})
	mb.Handle("/bar", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rr := safehttptest.NewTestResponseWriter(&b)

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/bar", nil)

	mux := mb.Mux()
	mux.ServeHTTP(rr, req)

	if err != nil {
		t.Fatalf("csp.Nonce: got error %v", err)
	}

	if nonce == "" {
		t.Fatalf("csp.Nonce: got %q, want non-empty", nonce)
	}

	if want, got := rr.Status(), safehttp.StatusOK; got != want {
		t.Errorf("rr.Status() got: %v want: %v", got, want)
	}

	wantHeaders := map[string][]string{
		"Content-Type": {"text/html; charset=utf-8"},
		"Content-Security-Policy": {
			"object-src 'none'; script-src 'unsafe-inline' 'nonce-" + nonce + "' 'strict-dynamic' https: http:; base-uri 'none'",
			"frame-ancestors 'self'"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header(): mismatch (-want +got):\n%s", diff)
	}

	wantBody := `<script nonce="` + nonce +
		`" type="application/javascript">alert("script")</script><h1>Content</h1>`
	if gotBody := b.String(); gotBody != wantBody {
		t.Errorf("response body: got %q, want nonce %q", gotBody, wantBody)
	}

}
