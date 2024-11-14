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

package csp

import (
	"strings"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp"
)

// TrustedTypesPolicy policy can be used to create a new CSP which makes
// dangerous web API functions secure by default.
//
// See https://web.dev/trusted-types for more info.
type TrustedTypesPolicy struct {
	// ReportURI controls the report-uri directive. If ReportUri is empty, no report-uri
	// directive will be set.
	ReportURI string
}

// Serialize serializes this policy for use in a Content-Security-Policy header
// or in a Content-Security-Policy-Report-Only header. A nonce will be provided
// to Serialize which can be used in 'nonce-{random-nonce}' values in directives.
func (t TrustedTypesPolicy) Serialize(nonce string, _ safehttp.InterceptorConfig) string {
	var b strings.Builder
	b.WriteString("require-trusted-types-for 'script'")

	if t.ReportURI != "" {
		b.WriteString("; report-uri ")
		b.WriteString(t.ReportURI)
	}

	return b.String()
}

// Match matches strict policies overrides.
func (TrustedTypesPolicy) Match(cfg safehttp.InterceptorConfig) bool {
	_, ok := cfg.(internalunsafecsp.DisableTrustedTypes)
	return ok
}

// Overridden checks the override level.
func (TrustedTypesPolicy) Overridden(cfg safehttp.InterceptorConfig) (disabled, reportOnly bool) {
	disable := cfg.(internalunsafecsp.DisableTrustedTypes)
	return disable.SkipReports, true
}
