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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/fetch_metadata"
	"github.com/google/go-safeweb/safehttp/safehttptest"
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
	tests := append(allowedRIPHeaders, allowedRIPNavHeaders...)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
		})
	}
}

func TestRejectedResourceIsolationEnforceMode(t *testing.T) {
	tests := append(disallowedRIPHeaders, disallowedRIPNavHeaders...)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusForbidden, safehttp.StatusCode(rec.Status()); want != got {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "Forbidden\n", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
		})
	}
}

type reportTests struct {
	name, method, site, mode, dest, report string
}

type methodLogger struct {
	report string
}

func (l *methodLogger) Log(r *safehttp.IncomingRequest) {
	l.report = r.Method()
}

func TestRejectedResourceIsolationEnforceModeWithLogger(t *testing.T) {
	var tests []reportTests
	for _, t := range disallowedRIPHeaders {
		tests = append(tests, reportTests{
			name:   t.name,
			method: t.method,
			site:   t.site,
			mode:   t.mode,
			dest:   t.dest,
			report: t.method,
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
			logger := &methodLogger{}
			p.Logger = logger
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusForbidden, safehttp.StatusCode(rec.Status()); want != got {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "Forbidden\n", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
			if test.report != logger.report {
				t.Errorf("logger.report: got %s, want %s", logger.report, test.report)
			}
		})
	}
}

func TestResourceIsolationReportMode(t *testing.T) {
	var tests []reportTests
	for _, t := range allowedRIPHeaders {
		tests = append(tests, reportTests{
			name:   t.name,
			method: t.method,
			site:   t.site,
			mode:   t.mode,
			dest:   t.dest,
			report: "",
		})
	}
	for _, t := range disallowedRIPHeaders {
		tests = append(tests, reportTests{
			name:   t.name,
			method: t.method,
			site:   t.site,
			mode:   t.mode,
			dest:   t.dest,
			report: t.method,
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
			logger := &methodLogger{}
			p.Logger = logger
			p.SetReportOnly()
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
			if logger.report != test.report {
				t.Errorf("logger.report: got %s, want %s", logger.report, test.report)
			}
		})
	}
}

func TestReportModeMissingLogger(t *testing.T) {
	p := fetchmetadata.NewPlugin()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Error("p.SetReportOnly(nil) expected panic")
	}()
	p.SetReportOnly()
}

func TestNavIsolationEnforceMode(t *testing.T) {
	tests := append(allowedRIPNavHeaders, disallowedRIPNavHeaders...)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin()
			p.NavIsolation = true
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusForbidden, safehttp.StatusCode(rec.Status()); want != got {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "Forbidden\n", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
		})
	}
}

func TestNavIsolationReportMode(t *testing.T) {
	type reportTests struct {
		name, method, site, mode, dest, report string
	}
	var tests []reportTests
	for _, t := range allowedRIPNavHeaders {
		tests = append(tests, reportTests{
			name:   t.name,
			method: t.method,
			site:   t.site,
			mode:   t.mode,
			dest:   t.dest,
			report: t.method,
		})
	}
	for _, t := range disallowedRIPNavHeaders {
		tests = append(tests, reportTests{
			name:   t.name,
			method: t.method,
			site:   t.site,
			mode:   t.mode,
			dest:   t.dest,
			report: t.method,
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
			logger := &methodLogger{}
			p.Logger = logger
			p.NavIsolation = true
			p.SetReportOnly()
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); want != got {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
			if logger.report != test.report {
				t.Errorf("logger.report: got %s, want %s", logger.report, test.report)
			}
		})
	}
}

func TestCORSEndpoint(t *testing.T) {
	tests := append(disallowedRIPHeaders, disallowedRIPNavHeaders...)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin("/carbonara")
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("rec.Body() got: %q want: %q", got, want)
			}
		})
	}
}

func TestCORSAfterRedirect(t *testing.T) {
	tests := []struct {
		name, method, site, mode, dest string
		wantHeaders                    map[string][]string
		wantBody                       string
	}{
		{
			name:   "cross origin POST",
			method: safehttp.MethodPost,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "document",
			wantHeaders: map[string][]string{
				"Location": {"https://spaghetti.com/carbonara"},
			},
			wantBody: "",
		},
		{
			name:   "cross origin GET from object",
			method: safehttp.MethodGet,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "object",
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Location":     {"https://spaghetti.com/carbonara"},
			},
			wantBody: "<a href=\"https://spaghetti.com/carbonara\">Moved Permanently</a>.\n\n",
		},
		{
			name:   "cross origin HEAD from embed",
			method: safehttp.MethodHead,
			site:   "cross-site",
			mode:   "navigate",
			dest:   "embed",
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Location":     {"https://spaghetti.com/carbonara"},
			},
			wantBody: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/bolognese", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			rec := safehttptest.NewResponseRecorder()

			p := fetchmetadata.NewPlugin("/carbonara")
			p.NavIsolation = true
			p.RedirectURL, _ = safehttp.ParseURL("https://spaghetti.com/carbonara")
			p.Before(rec.ResponseWriter, req, nil)

			if want, got := safehttp.StatusMovedPermanently, safehttp.StatusCode(rec.Status()); got != want {
				t.Errorf("rec.Status() got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(test.wantHeaders, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if gotBody := rec.Body(); test.wantBody != gotBody {
				t.Errorf("rec.Body() got: %q want: %q", gotBody, test.wantBody)
			}
		})
	}

}
