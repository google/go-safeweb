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
	"context"
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
			policy:     StrictCSPBuilder{}.Build(),
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http:; base-uri 'none'",
		},
		{
			name:       "StrictCSP with no strict-dynamic",
			policy:     StrictCSPBuilder{NoStrictDynamic: true}.Build(),
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret'; base-uri 'none'",
		},
		{
			name:       "StrictCSP with unsafe-eval",
			policy:     StrictCSPBuilder{UnsafeEval: true}.Build(),
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http: 'unsafe-eval'; base-uri 'none'",
		},
		{
			name:       "StrictCSP with set base-uri",
			policy:     StrictCSPBuilder{BaseURI: "https://example.com"}.Build(),
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http:; base-uri https://example.com",
		},
		{
			name:       "StrictCSP with report-uri",
			policy:     StrictCSPBuilder{ReportURI: "https://example.com/collector"}.Build(),
			wantString: "object-src 'none'; script-src 'unsafe-inline' 'nonce-super-secret' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
		},
		{
			name:       "FramingCSP",
			policy:     FramingPolicyBuilder{}.Build(),
			wantString: "frame-ancestors 'self'",
		},
		{
			name:       "FramingCSP with report-uri",
			policy:     FramingPolicyBuilder{ReportURI: "httsp://example.com/collector"}.Build(),
			wantString: "frame-ancestors 'self'; report-uri httsp://example.com/collector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.policy.serialize("super-secret")

			if s != tt.wantString {
				t.Errorf("tt.policy.serialize() got: %q want: %q", s, tt.wantString)
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
				"frame-ancestors 'self'",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:        "Default policies with reporting URI",
			interceptor: Default("https://example.com/collector"),
			wantEnforcePolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://example.com/collector",
				"frame-ancestors 'self'; report-uri https://example.com/collector",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "StrictCSP Report Only",
			interceptor: Interceptor{
				ReportOnly: []Policy{
					StrictCSPBuilder{ReportURI: "https://example.com/collector"}.Build(),
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
					FramingPolicyBuilder{ReportURI: "https://example.com/collector"}.Build(),
				},
			},
			wantReportOnlyPolicy: []string{"frame-ancestors 'self'; report-uri https://example.com/collector"},
			wantNonce:            "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := safehttptest.NewResponseRecorder()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			tt.interceptor.Before(rr.ResponseWriter, req)

			h := rr.Header()
			if diff := cmp.Diff(tt.wantEnforcePolicy, h.Values("Content-Security-Policy"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy\") mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantReportOnlyPolicy, h.Values("Content-Security-Policy-Report-Only"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy-Report-Only\") mismatch (-want +got):\n%s", diff)
			}

			v := req.Context().Value(ctxKey{})
			if v == nil {
				t.Fatalf("req.Context().Value(ctxKey{}) got: nil want: %q", tt.wantNonce)
			}
			if got := v.(string); got != tt.wantNonce {
				t.Errorf("v.(string) got: %q want: %q", got, tt.wantNonce)
			}
		})
	}
}

func TestAlreadyClaimed(t *testing.T) {
	headers := []string{
		"Content-Security-Policy",
		"Content-Security-Policy-Report-Only",
	}

	for _, h := range headers {
		t.Run(h, func(t *testing.T) {
			rr := safehttptest.NewResponseRecorder()
			if _, err := rr.ResponseWriter.Header().Claim(h); err != nil {
				t.Fatalf("rr.ResponseWriter.Header().Claim(h) got err: %v want: nil", err)
			}
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			it := Interceptor{}
			it.Before(rr.ResponseWriter, req)

			if got, want := rr.Status(), safehttp.StatusInternalServerError; got != want {
				t.Errorf("rr.Status() got: %v want: %v", got, want)
			}

			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}

			if got, want := rr.Body(), "Internal Server Error\n"; got != want {
				t.Errorf("rr.Body() got: %q want: %q", got, want)
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
		if r := recover(); r == nil {
			t.Error("generateNonce() expected panic")
		}
	}()
	generateNonce()
}

func TestValidNonce(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey{}, "nonce")
	n, err := Nonce(ctx)
	if err != nil {
		t.Errorf("Nonce(ctx) got err: %v want: nil", err)
	}

	if want := "nonce"; n != want {
		t.Errorf("Nonce(ctx) got nonce: %v want: %v", n, want)
	}
}

func TestNonceEmptyContext(t *testing.T) {
	ctx := context.Background()
	n, err := Nonce(ctx)
	if err == nil {
		t.Error("Nonce(ctx) got err: nil want: error")
	}

	if want := ""; n != want {
		t.Errorf("Nonce(ctx) got nonce: %v want: %v", n, want)
	}
}
