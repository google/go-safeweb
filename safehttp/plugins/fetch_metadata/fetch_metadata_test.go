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

func TestDefaultPolicy(t *testing.T) {
	tests := []struct {
		name    string
		req     *safehttp.IncomingRequest
		allowed bool
	}{
		{
			name:    `Sec-Fetch-Site: ""`,
			req:     safehttptest.NewRequest("POST", "/", nil),
			allowed: true,
		},
		{
			name: "Sec-Fetch-Site: same-origin",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "same-origin")
				return req
			}(),
			allowed: true,
		},
		{
			name: "Sec-Fetch-Site: same-site",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "same-site")
				return req
			}(),
			allowed: true,
		},
		{
			name: "Sec-Fetch-Site: none",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "none")
				return req
			}(),
			allowed: true,
		},
		{
			name: "Sec-Fetch-Site: cross-site",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				return req
			}(),
			allowed: false,
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
			allowed: true,
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
			allowed: false,
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
			allowed: false,
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
			allowed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := fetchmetadata.NewPlugin()
			got := p.Policy(test.req)

			if want := test.allowed; want != got {
				t.Errorf("p.Policy(test.req): got %v, want %v", got, want)
			}
		})
	}
}

func TestEnforceMode(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:        "Allowed Request",
			req:         safehttptest.NewRequest("POST", "/", nil),
			wantStatus:  safehttp.StatusOK,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Rejected Request",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "navigate")
				req.Header.Add("Sec-Fetch-Dest", "image")
				return req
			}(),
			wantStatus: safehttp.StatusForbidden,
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
			if status := safehttp.StatusCode(rec.Status()); status != test.wantStatus {
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

type fooLog struct {
	report string
}

func (l *fooLog) Log(r *safehttp.IncomingRequest) {
	l.report = r.Method() + " " + r.URL.Path()
}

func TestReportModeWithLogger(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  safehttp.StatusCode
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
			wantHeaders: map[string][]string{},
			wantBody:    "",
			wantReport:  "GET /pizza",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := fetchmetadata.NewPlugin()
			logger := &fooLog{}
			p.SetReportMode(logger)
			rec := safehttptest.NewResponseRecorder()

			p.Before(rec.ResponseWriter, test.req)

			if status := safehttp.StatusCode(rec.Status()); status != test.wantStatus {
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

func TestChangeMode(t *testing.T) {
	logger := &fooLog{}
	p := fetchmetadata.NewPlugin()
	req := safehttptest.NewRequest("POST", "/pizza", nil)
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	rec := safehttptest.NewResponseRecorder()

	p.Before(rec.ResponseWriter, req)

	if want, got := safehttp.StatusForbidden, safehttp.StatusCode(rec.Status()); want != got {
		t.Errorf("status code got: %v want: %v", got, want)
	}
	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}
	if want, got := "Forbidden\n", rec.Body(); got != want {
		t.Errorf("response body got: %q want: %q", got, want)
	}

	req = safehttptest.NewRequest("POST", "/pizza", nil)
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	rec = safehttptest.NewResponseRecorder()
	p.SetReportMode(logger)

	p.Before(rec.ResponseWriter, req)

	if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); want != got {
		t.Errorf("status code got: %v want: %v", got, want)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}
	if want, got := "", rec.Body(); got != want {
		t.Errorf("response body got: %q want: %q", got, want)
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
		wantStatus  safehttp.StatusCode
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus: safehttp.StatusForbidden,
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
			p.Policy = policy
			rec := safehttptest.NewResponseRecorder()

			p.Before(rec.ResponseWriter, test.req)

			if status := safehttp.StatusCode(rec.Status()); status != test.wantStatus {
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
