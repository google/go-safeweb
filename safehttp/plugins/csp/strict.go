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

// StrictPolicy can be used to build a strict, nonce-based CSP.
//
// See https://csp.withgoogle.com/docs/strict-csp.html for more info.
type StrictPolicy struct {
	// NoStrictDynamic controls whether script-src should contain the 'strict-dynamic'
	// value.
	//
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/script-src#strict-dynamic
	// for more info.
	NoStrictDynamic bool
	// UnsafeEval controls whether script-src should contain the 'unsafe-eval' value.
	// If enabled, the eval() JavaScript function is allowed.
	UnsafeEval bool
	// BaseURI controls the base-uri directive. If BaseURI is an empty string the
	// directive will be set to 'none'. The base-uri directive restricts the URLs
	// which can be used in a document's <base> element.
	//
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/base-uri
	// for more info.
	BaseURI string
	// ReportURI controls the report-uri directive. If ReportUri is empty, no report-uri
	// directive will be set.
	ReportURI string
	// Hashes adds a set of hashes to script-src. An example of a hash would be:
	//  sha256-CihokcEcBW4atb/CW/XWsvWwbTjqwQlE9nj9ii5ww5M=
	// which is the SHA256 hash for the script "console.log(1)".
	//
	// For more info, see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/script-src
	Hashes []string
}

// Serialize serializes this policy for use in a Content-Security-Policy header
// or in a Content-Security-Policy-Report-Only header. A nonce will be provided
// to Serialize which can be used in 'nonce-{random-nonce}' values in directives.
func (s StrictPolicy) Serialize(nonce string, _ safehttp.InterceptorConfig) string {
	var b strings.Builder

	// object-src 'none'; script-src 'unsafe-inline' 'nonce-{random}'
	b.WriteString("object-src 'none'; script-src 'unsafe-inline' 'nonce-")
	b.WriteString(nonce)
	b.WriteByte('\'')

	if !s.NoStrictDynamic {
		b.WriteString(" 'strict-dynamic' https: http:")
	}

	if s.UnsafeEval {
		b.WriteString(" 'unsafe-eval'")
	}

	for _, h := range s.Hashes {
		b.WriteString(" '")
		b.WriteString(h)
		b.WriteByte('\'')
	}

	b.WriteString("; base-uri ")
	if s.BaseURI == "" {
		b.WriteString("'none'")
	} else {
		b.WriteString(s.BaseURI)
	}

	if s.ReportURI != "" {
		b.WriteString("; report-uri ")
		b.WriteString(s.ReportURI)
	}

	return b.String()
}

// Match matches strict policies overrides.
func (StrictPolicy) Match(cfg safehttp.InterceptorConfig) bool {
	_, ok := cfg.(internalunsafecsp.DisableStrict)
	return ok
}

// Overridden checks the override level.
func (StrictPolicy) Overridden(cfg safehttp.InterceptorConfig) (disabled, reportOnly bool) {
	disable := cfg.(internalunsafecsp.DisableStrict)
	return disable.SkipReports, true
}
