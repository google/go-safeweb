// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fetchmetadata_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

var (
	allowedFIPHeaders = []testHeaders{
		{
			name: "Fetch Metadata not supported",
		},
		{
			name: "Same origin",
			site: "same-origin",
			mode: "navigate",
			dest: "frame",
		},
		{
			name: "User agent initiated",
			site: "none",
			mode: "navigate",
			dest: "frame",
		},
		{
			name: "Non-navigational",
			site: "cross-site",
			mode: "cors",
			dest: "frame",
		},
		{
			name: "Non-frameable",
			site: "cross-site",
			mode: "navigate",
			dest: "script",
		},
	}
	disallowedFIPHeaders = []testHeaders{
		{
			name: "Cross origin frame",
			site: "cross-origin",
			mode: "navigate",
			dest: "frame",
		},
		{
			name: "Same site, corss origin embed",
			site: "same-site",
			mode: "nested-navigate",
			dest: "embed",
		},
	}
)

func TestAllowedFramingIsolationEnforceMode(t *testing.T) {
	for _, test := range allowedFIPHeaders {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest("GET", "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			p := fetchmetadata.FramingIsolationPolicy()
			p.Before(fakeRW, req, nil)

			if want, got := int(safehttp.StatusOK), rr.Code; got != want {
				t.Errorf("rr.Code got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(http.Header{}, rr.Header()); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rr.Body.String(); got != want {
				t.Errorf("rr.Body.String() got: %q want: %q", got, want)
			}
		})
	}
}

func TestRejectedFramingIsolationEnforceMode(t *testing.T) {
	for _, test := range disallowedFIPHeaders {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest("GET", "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			p := fetchmetadata.FramingIsolationPolicy()
			p.Before(fakeRW, req, nil)

			if want, got := safehttp.StatusForbidden, safehttp.StatusCode(rr.Code); want != got {
				t.Errorf("rr.Code got: %v want: %v", got, want)
			}
			if diff := cmp.Diff(http.Header{}, rr.Header()); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDisableFramingIsolationPolicy(t *testing.T) {
	type reportTests struct {
		name, site, mode, dest, block string
	}
	var tests []reportTests
	for _, t := range allowedFIPHeaders {
		tests = append(tests, reportTests{
			name:  t.name,
			site:  t.site,
			mode:  t.mode,
			dest:  t.dest,
			block: "false",
		})
	}
	for _, t := range disallowedFIPHeaders {
		tests = append(tests, reportTests{
			name:  t.name,
			site:  t.site,
			mode:  t.mode,
			dest:  t.dest,
			block: "true",
		})
	}
	overrides := []struct {
		name  string
		value safehttp.InterceptorConfig
	}{
		{"disable", internalunsafeframing.Disable{SkipReports: true}},
		{"allowlist", internalunsafeframing.AllowList{}},
	}
	for _, override := range overrides {
		for _, test := range tests {
			t.Run(test.name+" "+override.name, func(t *testing.T) {
				req := safehttptest.NewRequest("GET", "https://spaghetti.com/carbonara", nil)
				req.Header.Add("Sec-Fetch-Site", test.site)
				req.Header.Add("Sec-Fetch-Mode", test.mode)
				req.Header.Add("Sec-Fetch-Dest", test.dest)
				fakeRW, rr := safehttptest.NewFakeResponseWriter()

				p := fetchmetadata.FramingIsolationPolicy()
				p.Before(fakeRW, req, override.value)

				if want, got := safehttp.StatusOK, safehttp.StatusCode(rr.Code); want != got {
					t.Errorf("rr.Code got: %v want: %v", got, want)
				}
				if diff := cmp.Diff(http.Header{}, rr.Header()); diff != "" {
					t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
				}
				if want, got := "", rr.Body.String(); got != want {
					t.Errorf("rr.Body.String() got: %q want: %q", got, want)
				}
			})
		}
	}
}
