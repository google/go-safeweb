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

package csp

import (
	"strings"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"
)

// FramingPolicy can be used to create a new CSP policy with frame-ancestors
// set to 'self'.
//
// To specify a list of allowed framing hostnames use interceptor configurations.
type FramingPolicy struct {
	// ReportURI controls the report-uri directive. If ReportUri is empty, no report-uri
	// directive will be set.
	ReportURI string
}

// Serialize serializes this policy for use in a Content-Security-Policy header
// or in a Content-Security-Policy-Report-Only header. A nonce will be provided
// to Serialize which can be used in 'nonce-{random-nonce}' values in directives.
func (f FramingPolicy) Serialize(nonce string, cfg safehttp.InterceptorConfig) string {
	var b strings.Builder

	var allow []string
	if a, ok := cfg.(internalunsafeframing.AllowList); ok {
		allow = a.Hostnames
	}
	b.WriteString(frameAncestors(allow))
	b.WriteString(report(f.ReportURI))

	return strings.TrimSpace(b.String())
}

// Match matches strict policies overrides.
func (FramingPolicy) Match(cfg safehttp.InterceptorConfig) bool {
	switch cfg.(type) {
	case internalunsafeframing.Disable, internalunsafeframing.AllowList:
		return true
	}
	return false
}

// Overridden checks the override level.
func (FramingPolicy) Overridden(cfg safehttp.InterceptorConfig) (disabled, reportOnly bool) {
	switch c := cfg.(type) {
	case internalunsafeframing.Disable:
		return c.SkipReports, true
	case internalunsafeframing.AllowList:
		return false, c.ReportOnly
	}
	// This should not happen.
	return false, false
}

func frameAncestors(sources []string) string {
	var b strings.Builder
	b.WriteString("frame-ancestors 'self'")

	for _, s := range sources {
		b.WriteString(" ")
		b.WriteString(s)
	}
	b.WriteString("; ")

	return b.String()
}
