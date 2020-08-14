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

package hsts_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/hsts"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestHSTS(t *testing.T) {
	var test = []struct {
		name        string
		plugin      hsts.Plugin
		req         *safehttp.IncomingRequest
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:       "HTTPS",
			plugin:     hsts.NewPlugin(),
			req:        safehttptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=63072000; includeSubDomains"}, // 63072000 seconds is two years
			},
			wantBody: "",
		},
		{
			name:       "HTTP",
			plugin:     hsts.NewPlugin(),
			req:        safehttptest.NewRequest("GET", "http://localhost/", nil),
			wantStatus: 301,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Location":     {"https://localhost/"},
			},
			wantBody: "<a href=\"https://localhost/\">Moved Permanently</a>.\n\n",
		},
		{
			name:       "HTTP behind proxy",
			plugin:     hsts.Plugin{BehindProxy: true},
			req:        safehttptest.NewRequest("GET", "http://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0; includeSubDomains"},
			},
			wantBody: "",
		},
		{
			name:       "Preload",
			plugin:     hsts.Plugin{Preload: true, DisableIncludeSubDomains: true},
			req:        safehttptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0; preload"},
			},
			wantBody: "",
		},
		{
			name:       "Preload and IncludeSubDomains",
			plugin:     hsts.Plugin{Preload: true},
			req:        safehttptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0; includeSubDomains; preload"},
			},
			wantBody: "",
		},
		{
			name:       "No preload and no includeSubDomains",
			plugin:     hsts.Plugin{DisableIncludeSubDomains: true},
			req:        safehttptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0"},
			},
			wantBody: "",
		},
		{
			name:       "Custom maxage",
			plugin:     hsts.Plugin{MaxAge: 3600 * time.Second},
			req:        safehttptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=3600; includeSubDomains"}, // 3600 seconds is 1 hour
			},
			wantBody: "",
		},
		{
			name:       "Negative maxage",
			plugin:     hsts.Plugin{MaxAge: -1 * time.Second},
			req:        safehttptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 500,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Internal Server Error\n",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			rr := safehttptest.NewResponseRecorder()

			tt.plugin.Before(rr.ResponseWriter, tt.req)

			if rr.Status() != tt.wantStatus {
				t.Errorf("status code got: %v want: %v", rr.Status(), tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := rr.Body(); got != tt.wantBody {
				t.Errorf("response body got: %q want: %q", got, tt.wantBody)
			}
		})
	}
}

func TestStrictTransportSecurityAlreadyImmutable(t *testing.T) {
	p := hsts.NewPlugin()

	req := safehttptest.NewRequest("GET", "https://localhost/", nil)

	rr := safehttptest.NewResponseRecorder()
	rr.ResponseWriter.Header().Claim("Strict-Transport-Security")

	p.Before(rr.ResponseWriter, req)

	if want := 500; rr.Status() != want {
		t.Errorf("status code got: %v want: %v", rr.Status(), want)
	}

	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}

	if got, want := rr.Body(), "Internal Server Error\n"; got != want {
		t.Errorf("response body got: %q want: %q", got, want)
	}
}
