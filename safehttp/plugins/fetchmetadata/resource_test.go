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
	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata/internalunsafefetchmetadata/unsafefetchmetadatafortests"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

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
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			p := fetchmetadata.ResourceIsolationPolicy()
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

func TestRejectedResourceIsolationEnforceMode(t *testing.T) {
	tests := append(disallowedRIPHeaders, disallowedRIPNavHeaders...)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			p := fetchmetadata.ResourceIsolationPolicy()
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

func TestDisableResourceIsolationPolicy(t *testing.T) {
	type reportTests struct {
		name, method, site, mode, dest, blocked string
	}
	var tests []reportTests
	for _, t := range append(allowedRIPNavHeaders, allowedRIPHeaders...) {
		tests = append(tests, reportTests{
			name:    t.name,
			method:  t.method,
			site:    t.site,
			mode:    t.mode,
			dest:    t.dest,
			blocked: "false",
		})
	}
	for _, t := range append(disallowedRIPNavHeaders, disallowedRIPHeaders...) {
		tests = append(tests, reportTests{
			name:    t.name,
			method:  t.method,
			site:    t.site,
			mode:    t.mode,
			dest:    t.dest,
			blocked: "true",
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(test.method, "https://spaghetti.com/carbonara", nil)
			req.Header.Add("Sec-Fetch-Site", test.site)
			req.Header.Add("Sec-Fetch-Mode", test.mode)
			req.Header.Add("Sec-Fetch-Dest", test.dest)
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			p := fetchmetadata.ResourceIsolationPolicy()
			p.Before(fakeRW, req, unsafefetchmetadatafortests.DisableResourceIsolationPolicy())

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
