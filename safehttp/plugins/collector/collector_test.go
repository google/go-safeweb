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

func TestValidReport(t *testing.T) {
	tests := []struct {
		name   string
		report string
		want   []collector.Report
	}{
		{
			name: "Custom report",
			report: `[{
				"type": "custom",
				"age": 10,
				"url": "https://example.com/vulnerable-page/",
				"userAgent": "chrome",
				"body": {
					"x": "y",
					"pizza": "hawaii",
					"roundness": 3.14
				}
			}]`,
			want: []collector.Report{
				collector.Report{
					Type:      "custom",
					Age:       10,
					URL:       "https://example.com/vulnerable-page/",
					UserAgent: "chrome",
					Body: map[string]interface{}{
						"x":         "y",
						"pizza":     "hawaii",
						"roundness": float64(3.14),
					},
				},
			},
		},
		{
			name: "Multiple reports",
			report: `[{
				"type": "custom",
				"age": 10,
				"url": "https://example.com/vulnerable-page/",
				"userAgent": "chrome",
				"body": {
					"x": "y",
					"pizza": "hawaii",
					"roundness": 3.14
				}
			},
			{
				"type": "custom",
				"age": 15,
				"url": "https://example.com/",
				"userAgent": "firefox",
				"body": {
					"x": "z",
					"pizza": "kebab",
					"roundness": 1.234
				}
			}]`,
			want: []collector.Report{
				collector.Report{
					Type:      "custom",
					Age:       10,
					URL:       "https://example.com/vulnerable-page/",
					UserAgent: "chrome",
					Body: map[string]interface{}{
						"x":         "y",
						"pizza":     "hawaii",
						"roundness": float64(3.14),
					},
				},
				collector.Report{
					Type:      "custom",
					Age:       15,
					URL:       "https://example.com/",
					UserAgent: "firefox",
					Body: map[string]interface{}{
						"x":         "z",
						"pizza":     "kebab",
						"roundness": float64(1.234),
					},
				},
			},
		},
		{
			name: "csp-violation",
			report: `[{
				"type": "csp-violation",
				"age": 10,
				"url": "https://example.com/vulnerable-page/",
				"userAgent": "chrome",
				"body": {
					"blockedURL": "https://evil.com/",
					"disposition": "enforce",
					"documentURL": "https://example.com/blah/blah",
					"effectiveDirective": "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
					"originalPolicy": "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
					"referrer": "https://example.com/",
					"sample": "alert(1)",
					"statusCode": 200,
					"sourceFile": "stuff.js",
					"lineNumber": 10,
					"columnNumber": 17
				}	
			}]`,
			want: []collector.Report{
				collector.Report{
					Type:      "csp-violation",
					Age:       10,
					URL:       "https://example.com/vulnerable-page/",
					UserAgent: "chrome",
					Body: collector.CSPReport{
						BlockedURL:         "https://evil.com/",
						Disposition:        "enforce",
						DocumentURL:        "https://example.com/blah/blah",
						EffectiveDirective: "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
						OriginalPolicy:     "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
						Referrer:           "https://example.com/",
						Sample:             "alert(1)",
						StatusCode:         200,
						ViolatedDirective:  "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
						SourceFile:         "stuff.js",
						LineNumber:         10,
						ColumnNumber:       17,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReports []collector.Report
			h := collector.Handler(func(r collector.Report) {
				gotReports = append(gotReports, r)
			}, func(r collector.CSPReport) {
				t.Fatalf("expected CSP reports handler not to be called")
			})

			req := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(tt.report))
			req.Header.Set("Content-Type", "application/reports+json")

			rr := safehttptest.NewResponseRecorder()
			h.ServeHTTP(rr.ResponseWriter, req)

			if diff := cmp.Diff(tt.want, gotReports); diff != "" {
				t.Errorf("reports gotten mismatch (-want +got):\n%s", diff)
			}

			if got, want := rr.Status(), safehttp.StatusNoContent; got != want {
				t.Errorf("rr.Status() got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
			if got, want := rr.Body(), ""; got != want {
				t.Errorf("rr.Body() got: %q want: %q", got, want)
			}
		})
	}
}

func TestValidDeprecatedCSPReport(t *testing.T) {
	tests := []struct {
		name   string
		report string
		want   collector.CSPReport
	}{
		{
			name: "Basic",
			report: `{
				"csp-report": {
					"blocked-uri": "https://evil.com/",
					"disposition": "enforce",
					"document-uri": "https://example.com/blah/blah",
					"effective-directive": "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
					"original-policy": "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
					"referrer": "https://example.com/",
					"script-sample": "alert(1)",
					"status-code": 200,
					"violated-directive": "script-src",
					"source-file": "stuff.js"
				}
			}`,
			want: collector.CSPReport{
				BlockedURL:         "https://evil.com/",
				Disposition:        "enforce",
				DocumentURL:        "https://example.com/blah/blah",
				EffectiveDirective: "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
				OriginalPolicy:     "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
				Referrer:           "https://example.com/",
				Sample:             "alert(1)",
				StatusCode:         200,
				ViolatedDirective:  "script-src",
				SourceFile:         "stuff.js",
			},
		},
		{
			name: "No csp-report key",
			report: `{
				"blocked-uri": "https://evil.com/",
				"disposition": "enforce",
				"document-uri": "https://example.com/blah/blah",
				"effective-directive": "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
				"original-policy": "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
				"referrer": "https://example.com/",
				"script-sample": "alert(1)",
				"status-code": 200,
				"violated-directive": "script-src",
				"source-file": "stuff.js"
			}`,
			want: collector.CSPReport{
				BlockedURL:         "https://evil.com/",
				Disposition:        "enforce",
				DocumentURL:        "https://example.com/blah/blah",
				EffectiveDirective: "script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
				OriginalPolicy:     "object-src 'none'; script-src sha256.eZcc0TlUfnQi64XLdiN5c/Vh2vtDbPaGtXRFyE7dRLo=",
				Referrer:           "https://example.com/",
				Sample:             "alert(1)",
				StatusCode:         200,
				ViolatedDirective:  "script-src",
				SourceFile:         "stuff.js",
			},
		},
		{
			name: "lineno and colno",
			report: `{
				"csp-report": {
					"lineno": 15,
					"colno": 10
				}
			}`,
			want: collector.CSPReport{
				LineNumber:   15,
				ColumnNumber: 10,
			},
		},
		{
			name: "line-number and column-number",
			report: `{
				"csp-report": {
					"line-number": 15,
					"column-number": 10
				}
			}`,
			want: collector.CSPReport{
				LineNumber:   15,
				ColumnNumber: 10,
			},
		},
		{
			name: "Both lineno and colno, and line-number and column-number",
			report: `{
				"csp-report": {
					"lineno": 7,
					"colno": 8,
					"line-number": 15,
					"column-number": 10
				}
			}`,
			want: collector.CSPReport{
				LineNumber:   7,
				ColumnNumber: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := collector.Handler(func(r collector.Report) {
				t.Fatalf("expected generic reports handler not to be called")
			}, func(r collector.CSPReport) {
				if diff := cmp.Diff(tt.want, r); diff != "" {
					t.Errorf("report mismatch (-want +got):\n%s", diff)
				}
			})

			req := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(tt.report))
			req.Header.Set("Content-Type", "application/csp-report")

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
		})
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
			name: "csp-report, invalid json",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(`{"a:"b"}`))
				r.Header.Set("Content-Type", "application/csp-report")
				return r
			}(),
			wantStatus: safehttp.StatusBadRequest,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Bad Request\n",
		},
		{
			name: "reports+json, invalid json",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(`[{"a:"b"}]`))
				r.Header.Set("Content-Type", "application/reports+json")
				return r
			}(),
			wantStatus: safehttp.StatusBadRequest,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Bad Request\n",
		},
		{
			name: "csp-report, valid json, csp-report is not an object",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(`{"csp-report":"b"}`))
				r.Header.Set("Content-Type", "application/csp-report")
				return r
			}(),
			wantStatus: safehttp.StatusBadRequest,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Bad Request\n",
		},
		{
			name: "reports+json, valid json, body is not an object",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(`[{
					"type": "xyz",
					"age": 10,
					"url": "https://example.com/",
					"userAgent": "chrome",
					"body": "not an object"
				}]`))
				r.Header.Set("Content-Type", "application/reports+json")
				return r
			}(),
			wantStatus: safehttp.StatusBadRequest,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Bad Request\n",
		},
		{
			name: "Negative uints",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/collector", strings.NewReader(`{
					"csp-report": {
						"status-code": -1,
						"lineno": -1,
						"colno": -1,
						"line-number": -1,
						"column-number": -1
					}
				}`))
				r.Header.Set("Content-Type", "application/csp-report")
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
				t.Errorf("expected collector not to be called")
			}, func(r collector.CSPReport) {
				t.Errorf("expected collector not to be called")
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
