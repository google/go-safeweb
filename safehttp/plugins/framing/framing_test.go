// Copyright 2022 Google LLC
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

package framing

import (
	"testing"

	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing/unsafeframing"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing/unsafeframingfortests"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

const headerKey = "X-Frame-Options"

func TestFramingProtection(t *testing.T) {
	req := safehttptest.NewRequest("GET", "https://go.dev/", nil)
	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	xfoInterceptor{}.Before(fakeRW, req, nil)
	if got, want := rr.Header().Get(headerKey), "SAMEORIGIN"; got != want {
		t.Errorf(`response.Header().Get(%q): got %q want %q`, headerKey, got, want)
	}
}

func TestDisableFramingProtection(t *testing.T) {
	tests := []struct {
		name   string
		config safehttp.InterceptorConfig
	}{
		{
			name:   "unsafeframing",
			config: unsafeframing.Disable("testing", true),
		},
		{
			name:   "allowlist",
			config: unsafeframing.Allow("testing", false, "test.goog"),
		},
		{
			name:   "internal",
			config: internalunsafeframing.Disable{SkipReports: true},
		},
		{
			name:   "testing",
			config: unsafeframingfortests.Disable(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := safehttptest.NewRequest("GET", "https://go.dev/", nil)
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			if got, want := (xfoInterceptor{}).Match(tt.config), true; got != want {
				t.Errorf(`xfoInterceptor{}.Match(%T): got %v, want %v`, tt.config, got, want)
			}

			xfoInterceptor{}.Before(fakeRW, req, tt.config)
			if got, want := rr.Header().Get(headerKey), "ALLOWALL"; got != want {
				t.Errorf(`response.Header().Get(%q): got %q want %q`, headerKey, got, want)
			}
		})
	}

}
