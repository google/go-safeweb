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

package fetchmetadata_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/fetch_metadata"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"testing"
)

func TestEnforceMode(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:        `Sec-Fetch-Site: ""`,
			req:         safehttptest.NewRequest("POST", "/", nil),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Sec-Fetch-Site: same-origin",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "same-origin")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Sec-Fetch-Site: same-site",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "same-site")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Sec-Fetch-Site: none",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "none")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Sec-Fetch-Site: cross-site",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				return req
			}(),
			wantStatus: 403,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name: "GET request with Sec-Fetch-Mode: navigate from image",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "image")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "POST request with Sec-Fetch-Mode: navigate from image",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "image")
				return req
			}(),
			wantStatus: 403,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name: "GET request with Sec-Fetch-Mode: navigate from object",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "object")
				return req
			}(),
			wantStatus: 403,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name: "GET request with Sec-Fetch-Mode: navigate from embed",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "embed")
				return req
			}(),
			wantStatus: 403,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := fetchmetadata.NewPlugin()
			rec := safehttptest.NewResponseRecorder()

			p.Before(rec.ResponseWriter, test.req)

			if status := rec.Status(); status != test.wantStatus {
				t.Errorf("status code got: %v want: %v", status, test.wantStatus)
			}
			if diff := cmp.Diff(test.wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if got := rec.Body(); got != test.wantBody {
				t.Errorf("response body got: %q want: %q", got, test.wantBody)
			}
		})
	}

}

func TestReportModeWithLogger(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
		wantReport  string
	}{
		{
			name: "GET request with Sec-Fetch-Mode: navigate from image",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "image")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
			wantReport:  "",
		},
		{
			name: "Sec-Fetch-Site: cross-site",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
			wantReport:  "POST /pizza",
		},
		{
			name: "POST request with Sec-Fetch-Mode: navigate from image",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "image")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
			wantReport:  "POST /pizza",
		},
		{
			name: "GET request with Sec-Fetch-Mode: navigate from object",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "object")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
			wantReport:  "GET /pizza",
		},
		{
			name: "GET request with Sec-Fetch-Mode: navigate from embed",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "embed")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
			wantReport:  "GET /pizza",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := &fooLog{}
			p := fetchmetadata.NewPlugin()
			p.SetReportMode(logger)
			rec := safehttptest.NewResponseRecorder()

			p.Before(rec.ResponseWriter, test.req)

			if status := rec.Status(); status != test.wantStatus {
				t.Errorf("status code got: %v want: %v", status, test.wantStatus)
			}
			if diff := cmp.Diff(test.wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if got := rec.Body(); got != test.wantBody {
				t.Errorf("response body got: %q want: %q", got, test.wantBody)
			}
			if logger.report != test.wantReport {
				t.Errorf("logger.report: got %s, want %s", logger.report, test.wantReport)
			}
		})
	}
}

func TestReportModeMissingLogger(t *testing.T) {
	p := fetchmetadata.NewPlugin()
	defer func() {
		if r := recover(); r == nil {
			t.Error("p.SetReportMode(nil) expected panic")
		}
	}()
	p.SetReportMode(nil)
}

type fooLog struct {
	report string
}

func (l *fooLog) Log(r *safehttp.IncomingRequest) {
	l.report = r.Method() + " " + r.URL.Path()
}

func TestChangeMode(t *testing.T) {
	logger := &fooLog{}
	p := fetchmetadata.NewPlugin()
	req := safehttptest.NewRequest("POST", "/pizza", nil)
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	rec := safehttptest.NewResponseRecorder()

	p.Before(rec.ResponseWriter, req)

	if status := rec.Status(); status != 403 {
		t.Errorf("status code got: %v want: %v", status, 403)
	}
	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}

	if wantBody, got := "Forbidden\n", rec.Body(); got != wantBody {
		t.Errorf("response body got: %q want: %q", got, wantBody)
	}

	req = safehttptest.NewRequest("POST", "/pizza", nil)
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	rec = safehttptest.NewResponseRecorder()
	p.SetReportMode(logger)

	p.Before(rec.ResponseWriter, req)

	if status := rec.Status(); status != 200 {
		t.Errorf("status code got: %v want: %v", status, 200)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}
	if wantBody, got := "", rec.Body(); got != wantBody {
		t.Errorf("response body got: %q want: %q", got, wantBody)
	}
	if want := "POST /pizza"; logger.report != want {
		t.Errorf("logger.report: got %s, want %s", logger.report, want)
	}
}

func TestCustomPolicy(t *testing.T) {
	policy := func(r *safehttp.IncomingRequest) bool {
		if r.Header.Get("Sec-Fetch-Mode") != "cors" {
			return false
		}
		return true
	}
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "Allowed request for custom policy",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Mode", "cors")
				return req
			}(),
			wantStatus:  200,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Rejected request for custom policy",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/pizza", nil)
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				return req
			}(),
			wantStatus: 403,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := fetchmetadata.NewPlugin()
			p.SetPolicy(policy)
			rec := safehttptest.NewResponseRecorder()

			p.Before(rec.ResponseWriter, test.req)

			if status := rec.Status(); status != test.wantStatus {
				t.Errorf("status code got: %v want: %v", status, test.wantStatus)
			}
			if diff := cmp.Diff(test.wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if got := rec.Body(); got != test.wantBody {
				t.Errorf("response body got: %q want: %q", got, test.wantBody)
			}
		})
	}

}
