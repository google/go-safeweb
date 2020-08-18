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
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"

	"github.com/google/go-cmp/cmp"
)

func readRand(b []byte) (int, error) {
	for i := range b {
		b[i] = byte(i)
	}
	return len(b), nil
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name       string
		policy     *Policy
		wantString string
		wantNonces map[Directive]string
	}{
		{
			name:       "Empty",
			policy:     &Policy{},
			wantString: "",
			wantNonces: map[Directive]string{},
		},
		{
			name: "Default",
			policy: func() *Policy {
				p := NewPolicy("https://foo.com/collector")
				p.readRand = readRand
				return p
			}(),
			wantString: "object-src 'none'; script-src 'nonce-AAECAwQFBgc=' 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://foo.com/collector",
			wantNonces: map[Directive]string{DirectiveScriptSrc: "AAECAwQFBgc="},
		},
		{
			name: "Empty Directive",
			policy: &Policy{
				Directives: []*PolicyDirective{
					{Directive: DirectiveScriptSrc, Values: nil, AddNonce: false},
				},
			},
			wantString: "script-src",
			wantNonces: map[Directive]string{},
		},
		{
			name: "No nonces",
			policy: &Policy{
				Directives: []*PolicyDirective{
					{Directive: DirectiveScriptSrc, Values: []string{ValueUnsafeEval}, AddNonce: false},
				},
			},
			wantString: "script-src 'unsafe-eval'",
			wantNonces: map[Directive]string{},
		},
		{
			name: "Two nonces",
			policy: func() *Policy {
				p := &Policy{
					Directives: []*PolicyDirective{
						{Directive: DirectiveScriptSrc, Values: []string{ValueUnsafeEval}, AddNonce: true},
						{Directive: DirectiveStyleSrc, Values: []string{ValueUnsafeEval}, AddNonce: true},
					},
				}
				p.readRand = readRand
				return p
			}(),
			wantString: "script-src 'nonce-AAECAwQFBgc=' 'unsafe-eval'; style-src 'nonce-AAECAwQFBgc=' 'unsafe-eval'",
			wantNonces: map[Directive]string{
				DirectiveScriptSrc: "AAECAwQFBgc=",
				DirectiveStyleSrc:  "AAECAwQFBgc=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, nonces := tt.policy.Serialize()

			if got := s; got != tt.wantString {
				t.Errorf("tt.policy.Serialize() got: %q want: %q", got, tt.wantString)
			}

			if diff := cmp.Diff(tt.wantNonces, nonces); diff != "" {
				t.Errorf("nonces mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBefore(t *testing.T) {
	tests := []struct {
		name                  string
		interceptor           Interceptor
		wantEnforcementPolicy []string
		wantReportPolicy      []string
		wantEnforcementNonces map[Directive]string
		wantReportNonces      map[Directive]string
	}{
		{
			name:        "No policies",
			interceptor: Interceptor{},
		},
		{
			name: "Default policy",
			interceptor: func() Interceptor {
				p := Default("https://foo.com/collector")
				p.EnforcementPolicy.readRand = readRand
				return p
			}(),
			wantEnforcementPolicy: []string{"object-src 'none'; script-src 'nonce-AAECAwQFBgc=' 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://foo.com/collector"},
			wantEnforcementNonces: map[Directive]string{DirectiveScriptSrc: "AAECAwQFBgc="},
		},
		{
			name: "Report",
			interceptor: Interceptor{
				ReportOnlyPolicy: &Policy{
					Directives: []*PolicyDirective{
						{Directive: DirectiveScriptSrc, Values: []string{ValueUnsafeInline}, AddNonce: false},
					},
				},
			},
			wantReportPolicy: []string{"script-src 'unsafe-inline'"},
		},
		{
			name: "Report with nonces",
			interceptor: Interceptor{
				ReportOnlyPolicy: func() *Policy {
					p := NewPolicy("https://foo.com/collector")
					p.readRand = readRand
					return p
				}(),
			},
			wantReportPolicy: []string{"object-src 'none'; script-src 'nonce-AAECAwQFBgc=' 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http:; base-uri 'none'; report-uri https://foo.com/collector"},
			wantReportNonces: map[Directive]string{DirectiveScriptSrc: "AAECAwQFBgc="},
		},
		{
			name: "Report and enforce",
			interceptor: Interceptor{
				ReportOnlyPolicy: &Policy{
					Directives: []*PolicyDirective{
						{Directive: DirectiveScriptSrc, Values: []string{ValueUnsafeInline}, AddNonce: false},
						{Directive: DirectiveBaseURI, Values: []string{ValueNone}, AddNonce: false},
					},
				},
				EnforcementPolicy: &Policy{
					Directives: []*PolicyDirective{
						{Directive: DirectiveScriptSrc, Values: []string{ValueUnsafeInline}, AddNonce: false},
					},
				},
			},
			wantEnforcementPolicy: []string{"script-src 'unsafe-inline'"},
			wantReportPolicy:      []string{"script-src 'unsafe-inline'; base-uri 'none'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := safehttptest.NewResponseRecorder()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

			tt.interceptor.Before(rr.ResponseWriter, req)

			h := rr.Header()
			if diff := cmp.Diff(tt.wantEnforcementPolicy, h.Values("Content-Security-Policy")); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy\") mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantReportPolicy, h.Values("Content-Security-Policy-Report-Only")); diff != "" {
				t.Errorf("h.Values(\"Content-Security-Policy-Report-Only\") mismatch (-want +got):\n%s", diff)
			}

			ctx := req.Context()
			en := ctx.Value(ctxKey("enforce"))
			if diff := cmp.Diff(tt.wantEnforcementNonces, en); !(en == nil && tt.wantEnforcementNonces == nil) && diff != "" {
				t.Errorf("ctx.Value(ctxKey(\"enforce\")) mismatch (-want +got):\n%s", diff)
			}

			rn := ctx.Value(ctxKey("report"))
			if diff := cmp.Diff(tt.wantReportNonces, rn); !(rn == nil && tt.wantReportNonces == nil) && diff != "" {
				t.Errorf("ctx.Value(ctxKey(\"report\")) mismatch (-want +got):\n%s", diff)
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
				t.Fatalf("rr.ResponseWriter.Header().Claim(h) got: %v want: nil", err)
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
