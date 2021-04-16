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

func TestHSTSReject(t *testing.T) {
	var test = []struct {
		name         string
		interceptor  hsts.Interceptor
		req          *safehttp.IncomingRequest
		wantStatus   safehttp.StatusCode
		wantBody     string
		wantRedirect string
	}{
		{
			name:         "HTTP",
			interceptor:  hsts.Default(),
			req:          safehttptest.NewRequest(safehttp.MethodGet, "http://localhost/", nil),
			wantStatus:   safehttp.StatusMovedPermanently,
			wantRedirect: "https://localhost/",
		},
		{
			name:        "Negative MaxAge",
			interceptor: hsts.Interceptor{MaxAge: -1 * time.Second},
			req:         safehttptest.NewRequest(safehttp.MethodGet, "https://localhost/", nil),
			wantStatus:  safehttp.StatusInternalServerError,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			tt.interceptor.Before(fakeRW, tt.req, nil)

			if gotStatus := rr.Code; gotStatus != int(tt.wantStatus) {
				t.Errorf("rr.Code got: %v want: %v", gotStatus, tt.wantStatus)
			}

			if got, want := fakeRW.RedirectURL, tt.wantRedirect; got != want {
				t.Errorf("RedirectURL got %q, want %q", got, want)
			}
		})
	}
}

func TestHSTSOK(t *testing.T) {
	var test = []struct {
		name         string
		interceptor  hsts.Interceptor
		req          *safehttp.IncomingRequest
		wantHeaders  map[string][]string
		wantRedirect string
	}{
		{
			name:        "HTTPS",
			interceptor: hsts.Default(),
			req:         safehttptest.NewRequest(safehttp.MethodGet, "https://localhost/", nil),
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=63072000; includeSubDomains"}, // 63072000 seconds is two years
			},
		},
		{
			name:        "HTTP behind proxy",
			interceptor: hsts.Interceptor{BehindProxy: true},
			req:         safehttptest.NewRequest(safehttp.MethodGet, "http://localhost/", nil),
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0; includeSubDomains"},
			},
		},
		{
			name:        "Preload",
			interceptor: hsts.Interceptor{Preload: true, DisableIncludeSubDomains: true},
			req:         safehttptest.NewRequest(safehttp.MethodGet, "https://localhost/", nil),
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0; preload"},
			},
		},
		{
			name:        "Preload and IncludeSubDomains",
			interceptor: hsts.Interceptor{Preload: true},
			req:         safehttptest.NewRequest(safehttp.MethodGet, "https://localhost/", nil),
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0; includeSubDomains; preload"},
			},
		},
		{
			name:        "No preload and no includeSubDomains",
			interceptor: hsts.Interceptor{DisableIncludeSubDomains: true},
			req:         safehttptest.NewRequest(safehttp.MethodGet, "https://localhost/", nil),
			wantHeaders: map[string][]string{
				// max-age=0 tells the browser to expire the HSTS protection.
				"Strict-Transport-Security": {"max-age=0"},
			},
		},
		{
			name:        "Custom MaxAge",
			interceptor: hsts.Interceptor{MaxAge: 3600 * time.Second},
			req:         safehttptest.NewRequest(safehttp.MethodGet, "https://localhost/", nil),
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=3600; includeSubDomains"}, // 3600 seconds is 1 hour
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			tt.interceptor.Before(fakeRW, tt.req, nil)

			if gotStatus := rr.Code; gotStatus != int(safehttp.StatusOK) {
				t.Errorf("rr.Code got: %v want: %v", gotStatus, safehttp.StatusOK)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
