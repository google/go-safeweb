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

package coop

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestBefore(t *testing.T) {
	type want struct {
		enf, rep []string
	}
	var tests = []struct {
		name                 string
		interceptor          Interceptor
		overrider            Overrider
		want, wantOverridden want
	}{
		{
			name:           "No policies, override on header",
			interceptor:    NewInterceptor(),
			overrider:      Override(Policy{Mode: SameOrigin}),
			wantOverridden: want{enf: []string{"same-origin"}},
		},
		{
			name:        "Default",
			interceptor: Default("coop"),
			want:        want{enf: []string{`same-origin; report-to "coop"`}},
		},
		{
			name: "policies, override disables enf",
			interceptor: NewInterceptor(Policy{
				Mode:           SameOriginAllowPopups,
				ReportingGroup: "coop-ap",
			}, Policy{
				Mode:           SameOrigin,
				ReportingGroup: "coop-so",
				ReportOnly:     true,
			},
			),
			overrider: Override(Policy{
				Mode:           SameOrigin,
				ReportingGroup: "coop-so",
				ReportOnly:     true,
			}),
			want: want{
				enf: []string{`same-origin-allow-popups; report-to "coop-ap"`},
				rep: []string{`same-origin; report-to "coop-so"`},
			},
			wantOverridden: want{
				rep: []string{`same-origin; report-to "coop-so"`},
			},
		},
		{
			name: "multiple RO",
			interceptor: NewInterceptor(Policy{
				Mode:           SameOriginAllowPopups,
				ReportingGroup: "coop-ap",
			}, Policy{
				Mode:           SameOrigin,
				ReportingGroup: "coop-so",
				ReportOnly:     true,
			}, Policy{
				Mode:           UnsafeNone,
				ReportingGroup: "coop-un",
				ReportOnly:     true,
			}),
			want: want{
				enf: []string{`same-origin-allow-popups; report-to "coop-ap"`},
				rep: []string{`same-origin; report-to "coop-so"`, `unsafe-none; report-to "coop-un"`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := func(rr *safehttptest.ResponseRecorder, w want) {
				t.Helper()
				h := rr.Header()
				enf, rep := h.Values("Cross-Origin-Opener-Policy"), h.Values("Cross-Origin-Opener-Policy-Report-Only")
				if diff := cmp.Diff(w.enf, enf); diff != "" {
					t.Errorf("Enforced COOP -want +got:\n%s", diff)
				}
				if diff := cmp.Diff(w.rep, rep); diff != "" {
					t.Errorf("Report Only COOP -want +got:\n%s", diff)
				}
				if rr.Status() != safehttp.StatusOK {
					t.Errorf("Status: got %v want: %v", rr.Status(), safehttp.StatusOK)
				}
				if rr.Body() != "" {
					t.Errorf("Got body: %q, didn't want one", rr.Body())
				}
			}
			// Non overridden
			{
				rr := safehttptest.NewResponseRecorder()
				req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
				tt.interceptor.Before(rr.ResponseWriter, req, nil)
				check(rr, tt.want)
			}
			// Overridden
			{
				rr := safehttptest.NewResponseRecorder()
				req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
				tt.interceptor.Before(rr.ResponseWriter, req, tt.overrider)
				check(rr, tt.wantOverridden)
			}
		})
	}
}
