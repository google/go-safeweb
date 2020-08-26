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

type testHeaders struct {
	name, method, site, mode, dest string
}

var (
	allowedRIPHeaders = []testHeaders{
		{
			name:   "Fetch Metadata not supported",
			method: safehttp.MethodGet,
			site:   "",
		},
		{
			name:   "same origin",
			method: safehttp.MethodGet,
			site:   "same-origin",
		},
		{
			name:   "same site",
			method: safehttp.MethodGet,
			site:   "same-site",
		},
		{
			name:   "user agent initiated",
			method: safehttp.MethodGet,
			site:   "none",
		},
		{
			name:   "cors bug missing mode",
			method: safehttp.MethodOptions,
			site:   "cross-site",
			mode:   "",
		},
	}

	allowedRIPNavHeaders = []testHeaders{
		{
			name:   "cross origin GET navigate from document",
			method: safehttp.MethodGet,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "document",
		},
		{
			name:   "cross origin HEAD navigate from document",
			method: safehttp.MethodHead,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "document",
		},
		{
			name:   "cross origin GET navigate from nested-document",
			method: safehttp.MethodGet,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "nested-document",
		},
		{
			name:   "cross origin HEAD navigate from nested-document",
			method: safehttp.MethodHead,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "nested-document",
		},
		{
			name:   "cross origin GET nested-navigate from document",
			method: safehttp.MethodGet,
			site:   "cross-site",
			mode:   "nested-navigate",
			dest:   "document",
		},
		{
			name:   "cross origin HEAD nested-navigate from document",
			method: safehttp.MethodHead,
			site:   "cross-site",
			mode:   "nested-navigate",
			dest:   "document",
		},
		{
			name:   "cross origin GET nested-navigate from nested-document",
			method: safehttp.MethodGet,
			site:   "cross-site",
			mode:   "nested-navigate",
			dest:   "nested-document",
		},
		{
			name:   "cross origin HEAD nested-navigate from nested-document",
			method: safehttp.MethodHead,
			site:   "cross-site",
			mode:   "nested-navigate",
			dest:   "nested-document",
		},
	}

	disallowedRIPNavHeaders = []testHeaders{
		{
			name:   "cross origin POST",
			method: safehttp.MethodPost,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "document",
		},
		{
			name:   "cross origin GET from object",
			method: safehttp.MethodGet,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "object",
		},
		{
			name:   "cross origin HEAD from embed",
			method: safehttp.MethodHead,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "embed",
		},
	}

	disallowedRIPHeaders = []testHeaders{
		{
			name:   "cross origin no cors",
			method: safehttp.MethodPost,
			site:   "cross-site",
			mode:   "cors",
			dest:   "document",
		},
		{
			name:   "cross origin no cors",
			method: safehttp.MethodPost,
			site:   "cross-site",
			mode:   "no-cors",
			dest:   "nested-document",
		},
	}
)

func TestAllowedResourceIsolationEnforceMode(t *testing.T) {
	tests := allowedRIPHeaders
	for _, test := range allowedRIPNavHeaders {
		tests = append(tests, testHeaders{
			name:   test.name,
			method: test.method,
			site:   test.site,
			mode:   test.mode,
			dest:   test.dest,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
			p.Before(rec.ResponseWriter, req)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("status code got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("response body got: %q want: %q", got, want)
			}
		})
	}
}

func TestRejectedResourceIsolationEnforceMode(t *testing.T) {
	tests := disallowedRIPHeaders
	for _, test := range disallowedRIPNavHeaders {
		tests = append(tests, testHeaders{
			name:   test.name,
			method: test.method,
			site:   test.site,
			mode:   test.mode,
			dest:   test.dest,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
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
		})
	}
}

type fooLog struct {
	report string
}

func (l *fooLog) Log(r *safehttp.IncomingRequest) {
	l.report = r.Method()
}
func TestAllowedResourceIsolationReportMode(t *testing.T) {
	tests := allowedRIPHeaders
	for _, test := range allowedRIPNavHeaders {
		tests = append(tests, testHeaders{
			name:   test.name,
			method: test.method,
			site:   test.site,
			mode:   test.mode,
			dest:   test.dest,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
			logger := &fooLog{}
			p.SetReportOnly(logger)
			p.Before(rec.ResponseWriter, req)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("status code got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("response body got: %q want: %q", got, want)
			}
			if want := ""; logger.report != want {
				t.Errorf("logger.report: got %s, want %s", logger.report, want)
			}
		})
	}
}

func TestRejectedResourceIsolationReportMode(t *testing.T) {
	tests := disallowedRIPHeaders
	for _, test := range disallowedRIPNavHeaders {
		tests = append(tests, testHeaders{
			name:   test.name,
			method: test.method,
			site:   test.site,
			mode:   test.mode,
			dest:   test.dest,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
			logger := &fooLog{}
			p.SetReportOnly(logger)
			p.Before(rec.ResponseWriter, req)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); want != got {
				t.Errorf("status code got: %v want: %v", got, want)
			}
			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("status code got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("response body got: %q want: %q", got, want)
			}
			if want := test.method; logger.report != want {
				t.Errorf("logger.report: got %s, want %s", logger.report, want)
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

func TestNavIsolation(t *testing.T) {
	tests := allowedRIPNavHeaders
	for _, test := range disallowedRIPNavHeaders {
		tests = append(tests, testHeaders{
			name:   test.name,
			method: test.method,
			site:   test.site,
			mode:   test.mode,
			dest:   test.dest,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
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
		})
	}
}

func TestCORSEndpoint(t *testing.T) {
	tests := disallowedRIPHeaders
	for _, test := range disallowedRIPNavHeaders {
		tests = append(tests, testHeaders{
			name:   test.name,
			method: test.method,
			site:   test.site,
			mode:   test.mode,
			dest:   test.dest,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin("https://spaghetti.com/carbonara")
			p.Before(rec.ResponseWriter, req)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("status code got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("response body got: %q want: %q", got, want)
			}
		})
	}
}
