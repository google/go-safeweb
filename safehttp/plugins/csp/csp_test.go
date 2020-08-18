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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

type dummyReader struct{}

func (dummyReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 41
	}
	return len(b), nil
}

func TestMain(m *testing.M) {
	randReader = dummyReader{}
	os.Exit(m.Run())
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name       string
		policy     Policy
		wantString string
		wantNonce  string
	}{
		{
			name:       "StrictCSP",
			policy:     NewStrictCSP(false, false, false, "", ""),
			wantString: "object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk='; base-uri 'none'",
			wantNonce:  "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:       "StrictCSP with strict-dynamic",
			policy:     NewStrictCSP(false, true, false, "", ""),
			wantString: "object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'strict-dynamic'; base-uri 'none'",
			wantNonce:  "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:       "StrictCSP with unsafe-eval",
			policy:     NewStrictCSP(false, false, true, "", ""),
			wantString: "object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'unsafe-eval'; base-uri 'none'",
			wantNonce:  "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:       "StrictCSP with set base-uri",
			policy:     NewStrictCSP(false, false, false, "https://example.com", ""),
			wantString: "object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk='; base-uri https://example.com",
			wantNonce:  "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:       "StrictCSP with report-uri",
			policy:     NewStrictCSP(false, false, true, "", "https://example.com/collector"),
			wantString: "object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk=' 'unsafe-eval'; base-uri 'none'; report-uri https://example.com/collector",
			wantNonce:  "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:       "FramingCSP",
			policy:     NewFramingCSP(false),
			wantString: "frame-ancestors 'self'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := tt.policy.serialize(context.Background())

			if s != tt.wantString {
				t.Errorf("tt.policy.serialize() got: %q want: %q", s, tt.wantString)
			}

			v := ctx.Value(ctxKey{})
			if v == nil {
				v = ""
			}
			if got := v.(string); got != tt.wantNonce {
				t.Errorf("ctx.Value(ctxKey{}) got: %q want: %q", got, tt.wantNonce)
			}
		})
	}
}

func TestBefore(t *testing.T) {
	tests := []struct {
		name                  string
		interceptor           Interceptor
		wantEnforcementPolicy []string
		wantReportOnlyPolicy  []string
		wantNonce             string
	}{
		{
			name:        "No policies",
			interceptor: Interceptor{},
		},
		{
			name:        "Default policies",
			interceptor: Default(""),
			wantEnforcementPolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk='; base-uri 'none'",
				"frame-ancestors 'self'",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:        "Default policies with reporting URI",
			interceptor: Default("https://example.com/collector"),
			wantEnforcementPolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk='; base-uri 'none'; report-uri https://example.com/collector",
				"frame-ancestors 'self'",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name: "StrictCSP reportonly",
			interceptor: Interceptor{
				Policies: []Policy{
					NewStrictCSP(true, false, false, "", "https://example.com/collector"),
				},
			},
			wantReportOnlyPolicy: []string{
				"object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-KSkpKSkpKSkpKSkpKSkpKSkpKSk='; base-uri 'none'; report-uri https://example.com/collector",
			},
			wantNonce: "KSkpKSkpKSkpKSkpKSkpKSkpKSk=",
		},
		{
			name:                 "FramingCSP reportonly",
			interceptor:          Interceptor{Policies: []Policy{NewFramingCSP(true)}},
			wantReportOnlyPolicy: []string{"frame-ancestors 'self'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := safehttptest.NewResponseRecorder()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			tt.interceptor.Before(rr.ResponseWriter, req)

			h := rr.Header()
			if diff := cmp.Diff(tt.wantEnforcementPolicy, h.Values("Content-Security-Policy"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy\") mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantReportOnlyPolicy, h.Values("Content-Security-Policy-Report-Only"), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy-Report-Only\") mismatch (-want +got):\n%s", diff)
			}

			ctx := req.Context()
			v := ctx.Value(ctxKey{})
			if v == nil {
				v = ""
			}
			if got := v.(string); got != tt.wantNonce {
				t.Errorf("ctx.Value(ctxKey{}) got: %q want: %q", got, tt.wantNonce)
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

			if got, want := rr.Status(), int(safehttp.StatusInternalServerError); got != want {
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
