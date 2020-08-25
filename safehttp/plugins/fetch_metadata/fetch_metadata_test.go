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

func TestResourceIsolationEnforceMode(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:        `Sec-Fetch-Site: ""`,
			req:         safehttptest.NewRequest("POST", "/", nil),
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus:  safehttp.StatusOK,
			wantHeaders: map[string][]string{},
			wantBody:    "",
		},
		{
			name: "Sec-Fetch-Site: cross-site non-navigational",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Site", "no-cors")
				return req
			}(),
			wantStatus: safehttp.StatusForbidden,
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
			wantStatus:  safehttp.StatusOK,
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
			wantStatus: safehttp.StatusForbidden,
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
			wantStatus: safehttp.StatusForbidden,
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

func TestResourceIsolationReportMode(t *testing.T) {
	tests := []struct {
		name        string
		req         *safehttp.IncomingRequest
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
		wantReport  string
	}{
		{
			name: "Sec-Fetch-Site: cross-site",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "https://spaghetti.com/pizza", nil)
				req.Header.Add("Sec-Fetch-Site", "cross-site")
				req.Header.Add("Sec-Fetch-Mode", "cors")
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
				req := safehttptest.NewRequest("POST", "https://spaghetti.com/pizza", nil)
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
				req := safehttptest.NewRequest("GET", "https://spaghetti.com/pizza", nil)
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
				req := safehttptest.NewRequest("GET", "https://spaghetti.com/pizza", nil)
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
		{
			name: "GET request with Sec-Fetch-Mode: navigate from image",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("GET", "https://spaghetti.com/pizza", nil)
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := fetchmetadata.NewPlugin()
			logger := &fooLog{}
			p.SetReportOnly(logger)
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
			t.Error("p.SetReportOnly(nil) expected panic")
		}
	}()
	p.SetReportOnly(nil)
}

func TestChangeEnforceReportMode(t *testing.T) {
	logger := &fooLog{}
	p := fetchmetadata.NewPlugin()
	req := safehttptest.NewRequest("POST", "https://spaghetti.com/pizza", nil)
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Dest", "image")
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

	rec = safehttptest.NewResponseRecorder()
	p.SetReportOnly(logger)

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

func TestEnableDisableNavIsolation(t *testing.T) {
	req := safehttptest.NewRequest("GET", "https://spaghetti.com/pizza", nil)
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Dest", "image")
	rec := safehttptest.NewResponseRecorder()

	p := fetchmetadata.NewPlugin()
	p.NavIsolation = true
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

	rec = safehttptest.NewResponseRecorder()

	p.NavIsolation = false
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
}

func TestCORSEndpoint(t *testing.T) {
	methods := []string{"GET", "POST"}
	for _, m := range methods {
		req := safehttptest.NewRequest(m, "https://spaghetti.com/pizza", nil)
		req.Header.Add("Sec-Fetch-Site", "cross-site")
		req.Header.Add("Sec-Fetch-Mode", "navigate")
		req.Header.Add("Sec-Fetch-Dest", "image")
		rec := safehttptest.NewResponseRecorder()
		p := fetchmetadata.NewPlugin("https://spaghetti.com/pizza")
		p.NavIsolation = true
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
	}
}
