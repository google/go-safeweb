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

package csp

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

type endlessAReader struct{}

func (endlessAReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 41
	}
	return len(b), nil
}

func TestMain(m *testing.M) {
	randReader = endlessAReader{}
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
			name: "FramingCSP with one source",
			policy: FramingPolicy{Sources: []string{
				"https://www.example.org",
			}},
			wantString: "frame-ancestors 'self' https://www.example.org;",
		},
		{
			name: "FramingCSP with multiple sources",
			policy: FramingPolicy{Sources: []string{
				"https://a.example.org",
				"https://b.example.org",
			}},
			wantString: "frame-ancestors 'self' https://a.example.org https://b.example.org;",
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
			s := tt.policy.Serialize("super-secret")

			if s != tt.wantString {
				t.Errorf("tt.policy.Serialize() got: %q want: %q", s, tt.wantString)
			}
		})
	}
}

func TestBefore(t *testing.T) {
	tests := []struct {
		name                 string
		interceptor          Interceptor
		wantEnforcePolicy    []string
		wantReportOnlyPolicy []string
		wantNonce            string
	}{
		{
			name:        "No policies",
			interceptor: Interceptor{},
			wantNonce:   "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:        "Default policies",
			interceptor: Default(""),
			wantEnforcePolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'",
				"frame-ancestors 'self';",
				"require-trusted-types-for 'script'",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:        "Default policies with reporting URI",
			interceptor: Default("https://example.com/collector"),
			wantEnforcePolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
				"frame-ancestors 'self'; report-uri https://example.com/collector;",
				"require-trusted-types-for 'script'; report-uri https://example.com/collector",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "StrictCSP Report Only",
			interceptor: Interceptor{
				ReportOnly: []Policy{
					StrictPolicy{ReportURI: "https://example.com/collector"},
				},
			},
			wantReportOnlyPolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "FramingCSP Report Only",
			interceptor: Interceptor{
				ReportOnly: []Policy{
					FramingPolicy{ReportURI: "https://example.com/collector"},
				},
			},
			wantReportOnlyPolicy: []string{"frame-ancestors 'self'; report-uri https://example.com/collector;"},
			wantNonce:            "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			tt.interceptor.Before(fakeRW, req, nil)

			h := rr.Header()
			if diff := cmp.Diff(tt.wantEnforcePolicy, h.Values("Content-Security-Policy"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy\") mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantReportOnlyPolicy, h.Values("Content-Security-Policy-Report-Only"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy-Report-Only\") mismatch (-want +got):\n%s", diff)
			}

			v := safehttp.FlightValues(req.Context()).Get(nonceKey{})
			if v == nil {
				t.Fatalf("FlightValues(...).Get(nonceKey{}) got: nil want: %q", tt.wantNonce)
			}
			if got := v.(string); got != tt.wantNonce {
				t.Errorf("v.(string) got: %q want: %q", got, tt.wantNonce)
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read(b []byte) (int, error) {
	return 0, errors.New("bad")
}

func TestPanicWhileGeneratingNonce(t *testing.T) {
	randReader = errorReader{}
	defer func() {
		// TODO: avoid replacing the randReader per individual test cases
		randReader = endlessAReader{}
	}()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Error("generateNonce() expected panic")
	}()
	generateNonce()
}

func TestValidNonce(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	safehttp.FlightValues(req.Context()).Put(nonceKey{}, "nonce")
	n, err := Nonce(req.Context())
	if err != nil {
		t.Errorf("Nonce(ctx) got err: %v want: nil", err)
	}

	if want := "nonce"; n != want {
		t.Errorf("Nonce(ctx) got nonce: %v want: %v", n, want)
	}
}

func TestNonceEmptyContext(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	safehttp.FlightValues(req.Context()) // not putting nonce
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
	safehttp.FlightValues(req.Context()).Put(nonceKey{}, "pizza")

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
	safehttp.FlightValues(req.Context()).Put(nonceKey{}, nil /* missing nonce */)

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
