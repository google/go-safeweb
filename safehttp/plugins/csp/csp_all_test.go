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

package csp

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp"
	"github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp/unsafecspfortests"
	"github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp/unsafestrictcsp"
	"github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp/unsafetrustedtypes"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing/unsafeframing"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestMain(m *testing.M) {
	unsafecspfortests.UseStaticRandom()
	os.Exit(m.Run())
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name       string
		policy     Policy
		wantString string
	}{
		{
			name:       "StrictCSP",
			policy:     StrictPolicy{},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http:; base-uri 'none'",
		},
		{
			name:       "StrictCSP with no strict-dynamic",
			policy:     StrictPolicy{NoStrictDynamic: true},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret'; base-uri 'none'",
		},
		{
			name:       "StrictCSP with unsafe-eval",
			policy:     StrictPolicy{UnsafeEval: true},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http: 'unsafe-eval'; base-uri 'none'",
		},
		{
			name:       "StrictCSP with set base-uri",
			policy:     StrictPolicy{BaseURI: "https://example.com"},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http:; base-uri https://example.com",
		},
		{
			name:       "StrictCSP with report-uri",
			policy:     StrictPolicy{ReportURI: "https://example.com/collector"},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
		},
		{
			name: "StrictCSP with one hash",
			policy: StrictPolicy{Hashes: []string{
				"sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M=",
			}},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http: 'sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M='; base-uri 'none'",
		},
		{
			name: "StrictCSP with multiple hashes",
			policy: StrictPolicy{Hashes: []string{
				"sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M=",
				"sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M=",
			}},
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http: 'sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M=' 'sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M='; base-uri 'none'",
		},
		{
			name:       "FramingCSP",
			policy:     FramingPolicy{},
			wantString: "frame-ancestors 'self';",
		},
		{
			name:       "FramingCSP with report-uri",
			policy:     FramingPolicy{ReportURI: "httsp://example.com/collector"},
			wantString: "frame-ancestors 'self'; report-uri httsp://example.com/collector;",
		},
		{
			name:       "TrustedTypesCSP",
			policy:     TrustedTypesPolicy{},
			wantString: "require-trusted-types-for 'script'",
		},
		{
			name:       "TrustedTypesCSP with report-uri",
			policy:     TrustedTypesPolicy{ReportURI: "httsp://example.com/collector"},
			wantString: "require-trusted-types-for 'script'; report-uri httsp://example.com/collector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.policy.Serialize("super-secret", nil)

			if s != tt.wantString {
				t.Errorf("tt.policy.Serialize() got: %q want: %q", s, tt.wantString)
			}
		})
	}
}

func TestBefore(t *testing.T) {
	tests := []struct {
		name                 string
		interceptors         []Interceptor
		wantEnforcePolicy    []string
		wantReportOnlyPolicy []string
		wantNonce            string
	}{
		{
			name:         "Default policies",
			interceptors: Default(""),
			wantEnforcePolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'",
				"require-trusted-types-for 'script'",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:         "All policies",
			interceptors: append(Default(""), Interceptor{Policy: FramingPolicy{}}),
			wantEnforcePolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'",
				"require-trusted-types-for 'script'",
				"frame-ancestors 'self';",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "All policies with reporting URI",
			interceptors: append(Default("https://example.com/collector"),
				Interceptor{Policy: FramingPolicy{ReportURI: "https://example.com/collector"}}),
			wantEnforcePolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
				"require-trusted-types-for 'script'; report-uri https://example.com/collector",
				"frame-ancestors 'self'; report-uri https://example.com/collector;",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "StrictCSP Report Only",
			interceptors: []Interceptor{{
				Policy:     StrictPolicy{ReportURI: "https://example.com/collector"},
				ReportOnly: true,
			}},
			wantReportOnlyPolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "FramingCSP Report Only",
			interceptors: []Interceptor{{
				Policy:     FramingPolicy{ReportURI: "https://example.com/collector"},
				ReportOnly: true,
			}},
			wantReportOnlyPolicy: []string{"frame-ancestors 'self'; report-uri https://example.com/collector;"},
			wantNonce:            "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			for _, i := range tt.interceptors {
				i.Before(fakeRW, req, nil)
			}

			h := rr.Header()
			if diff := cmp.Diff(tt.wantEnforcePolicy, h.Values("Content-Security-Policy"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy\") mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantReportOnlyPolicy, h.Values("Content-Security-Policy-Report-Only"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy-Report-Only\") mismatch (-want +got):\n%s", diff)
			}

			v := safehttp.FlightValues(req.Context()).Get(nonceKey)
			if v == nil {
				t.Fatalf("safehttp.FlightValues(req.Context()).Get(nonceCtxKey) got: nil want: %q", tt.wantNonce)
			}
			if got := v.(string); got != tt.wantNonce {
				t.Errorf("v.(string) got: %q want: %q", got, tt.wantNonce)
			}
		})
	}
}

func TestValidNonce(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	_ = nonce(req)

	n, err := Nonce(req.Context())
	if err != nil {
		t.Errorf("Nonce(ctx) got err: %v want: nil", err)
	}

	if want := "KSkpKSkpKSkpKSkpKSkpKSkpKSk="; n != want {
		t.Errorf("Nonce(ctx) got nonce: %v want: %v", n, want)
	}
}

func TestNonceEmptyContext(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	// Not using nonce() to insert the nonce in context.

	n, err := Nonce(req.Context())
	if err == nil {
		t.Error("Nonce(ctx) got err: nil want: error")
	}

	if want := ""; n != want {
		t.Errorf("Nonce(ctx) got nonce: %v want: %v", n, want)
	}
}

func TestCommitNonce(t *testing.T) {
	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	safehttp.FlightValues(req.Context()).Put(nonceKey, "pizza")

	it := Interceptor{}
	tr := &safehttp.TemplateResponse{}
	it.Commit(fakeRW, req, tr, nil)

	nonce, ok := tr.FuncMap["CSPNonce"]
	if !ok {
		t.Fatal(`tr.FuncMap["CSPNonce"] not found`)
	}

	fn, ok := nonce.(func() string)
	if !ok {
		t.Fatalf(`tr.FuncMap["CSPNonce"]: got %T, want "func() string"`, fn)
	}
	if got, want := fn(), "pizza"; want != got {
		t.Errorf(`tr.FuncMap["CSPNonce"](): got %q, want %q`, got, want)
	}

	if got, want := rr.Code, int(safehttp.StatusOK); got != want {
		t.Errorf("rr.Code: got %v, want %v", got, want)
	}

	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}

	if got, want := "", rr.Body.String(); got != want {
		t.Errorf("rr.Body.String(): got %q want %q", got, want)
	}
}

func TestCommitMissingNonce(t *testing.T) {
	fakeRW, _ := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	// Not adding safehttp.FlightValues here.

	it := Interceptor{}
	tr := &safehttp.TemplateResponse{}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	it.Commit(fakeRW, req, tr, nil)
}

func TestCommitNotTemplateResponse(t *testing.T) {
	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	it := Interceptor{}
	it.Commit(fakeRW, req, safehttp.NoContentResponse{}, nil)

	if got, want := rr.Code, int(safehttp.StatusOK); got != want {
		t.Errorf("rr.Code: got %v, want %v", got, want)
	}

	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}

	if got, want := rr.Body.String(), ""; got != want {
		t.Errorf("rr.Body.String(): got %q want %q", got, want)
	}

}

func TestOverride(t *testing.T) {
	tests := []struct {
		name                 string
		interceptors         []Interceptor
		overrides            []safehttp.InterceptorConfig
		wantEnforcePolicy    []string
		wantReportOnlyPolicy []string
	}{
		{
			name:         "All policies, completely disabled",
			interceptors: append(Default(""), Interceptor{Policy: FramingPolicy{}}),
			overrides: []safehttp.InterceptorConfig{
				internalunsafecsp.DisableStrict{SkipReports: true},
				internalunsafecsp.DisableTrustedTypes{SkipReports: true},
				internalunsafeframing.Disable{SkipReports: true},
			},
		},
		{
			name:         "All policies, disabled via unsafe packages",
			interceptors: append(Default(""), Interceptor{Policy: FramingPolicy{}}),
			overrides: []safehttp.InterceptorConfig{
				unsafestrictcsp.Disable("testing", true),
				unsafetrustedtypes.Disable("testing", true),
				unsafeframing.Disable("testing", true),
			},
		},
		{
			name:         "All policies, report-only override",
			interceptors: append(Default(""), Interceptor{Policy: FramingPolicy{}}),
			overrides: []safehttp.InterceptorConfig{
				internalunsafecsp.DisableStrict{},
				internalunsafecsp.DisableTrustedTypes{},
				internalunsafeframing.Disable{},
			},
			wantReportOnlyPolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'",
				"require-trusted-types-for 'script'",
				"frame-ancestors 'self';",
			},
		},
		{
			name: "FramingCSP allowlist",
			interceptors: []Interceptor{Interceptor{
				Policy: FramingPolicy{}}},
			overrides: []safehttp.InterceptorConfig{
				unsafeframing.Allow("testing", true, "https://www.example.org"),
			},
			wantReportOnlyPolicy: []string{"frame-ancestors 'self' https://www.example.org;"},
		},
		{
			name: "FramingCSP allowlist",
			interceptors: []Interceptor{Interceptor{
				Policy: FramingPolicy{}}},
			overrides: []safehttp.InterceptorConfig{
				unsafeframing.Allow("testing", false, "https://a.example.org", "https://b.example.org"),
			},
			wantEnforcePolicy: []string{
				"frame-ancestors 'self' https://a.example.org https://b.example.org;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			for _, i := range tt.interceptors {
				var cfg safehttp.InterceptorConfig
				for _, c := range tt.overrides {
					if i.Match(c) {
						if cfg != nil {
							t.Fatalf("Multiple overrides match: %v and %v", cfg, c)
						}
						cfg = c
					}
				}
				i.Before(fakeRW, req, cfg)
			}

			h := rr.Header()
			if diff := cmp.Diff(tt.wantEnforcePolicy, h.Values("Content-Security-Policy"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy\") mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantReportOnlyPolicy, h.Values("Content-Security-Policy-Report-Only"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy-Report-Only\") mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
