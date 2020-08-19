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

package collector_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/collector"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestReport(t *testing.T) {
	report := `{
		"blocked-uri": "https://evil.com/",
		"disposition": "enforce",
		"document-uri": "https://example.com/blah/blah",
		"effective-directive": "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
		"original-policy": "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
		"referrer": "https://example.com/",
		"script-sample": "alert(1)",
		"status-code": 200,
		"violated-directive": "script-src"
	}`
	want := collector.Report{
		BlockedURI:         "https://evil.com/",
		Disposition:        "enforce",
		DocumentURI:        "https://example.com/blah/blah",
		EffectiveDirective: "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
		OriginalPolicy:     "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
		Referrer:           "https://example.com/",
		ScriptSample:       "alert(1)",
		StatusCode:         200,
		ViolatedDirective:  "script-src",
	}

	h := collector.Handler(func(r collector.Report) {
		if diff := cmp.Diff(want, r); diff != "" {
			t.Errorf("report mismatch (-want +got):\n%s", diff)
		}
	})

	req := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(report))
	req.Header.Set("Content-Type", "application/json")

	rr := safehttptest.NewResponseRecorder()
	h.ServeHTTP(rr.ResponseWriter, req)

	if got, want := rr.Status(), safehttp.StatusNoContent; got != want {
		t.Errorf("rr.Status() got: %v want: %v", got, want)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
	if got, want := rr.Body(), ""; got != want {
		t.Errorf("rr.Body() got: %q want: %q", got, want)
	}
}

func TestInvalidRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:       "Method",
			req:        safehttptest.NewRequest(safehttp.MethodGet, "/collector", nil),
			wantStatus: safehttp.StatusMethodNotAllowed,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Method Not Allowed\n",
		},
		{
			name: "Content-Type",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", nil)
				r.Header.Set("Content-Type", "text/plain")
				return r
			}(),
			wantStatus: safehttp.StatusUnsupportedMediaType,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Unsupported Media Type\n",
		},
		{
			name: "Body",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(`{"a:"b"}`))
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			wantStatus: safehttp.StatusBadRequest,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Bad Request\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := collector.Handler(func(r collector.Report) {
				t.Errorf("expected collector to not be called")
			})

			rr := safehttptest.NewResponseRecorder()
			h.ServeHTTP(rr.ResponseWriter, tt.req)

			if got := rr.Status(); got != tt.wantStatus {
				t.Errorf("rr.Status() got: %v want: %v", got, tt.wantStatus)
			}
			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
			if got := rr.Body(); got != tt.wantBody {
				t.Errorf("rr.Body() got: %q want: %q", got, tt.wantBody)
			}
		})
	}
}
