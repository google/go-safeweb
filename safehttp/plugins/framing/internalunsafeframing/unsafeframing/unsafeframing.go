// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package unsafeframing can be used to disable Framing protections on specific handler registration.
//
// Usage of this package should require a security review.
package unsafeframing

import (
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"
)

// Disable turns framing protections to report-only where supported, otherwise turns them off.
// If skipReports is true, all protections will be turned completely off.
func Disable(reason string, skipReports bool) internalunsafeframing.Disable {
	if reason == "" {
		panic("reason cannot be empty")
	}
	return internalunsafeframing.Disable{SkipReports: skipReports}
}

// Allow permits to specify a set of hostnames (with potential wildcards) that will be able to frame the site.
//
// Wildcards must follow the CSP specification:
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-ancestors.
//
// If reportOnly is true the policy will be set to Report-Only, which provides not security benefit
// but can be used to detect potential breakages.
//
// Please note that this option is only supported by browsers that support CSP: older browsers
// will end up allowing all origins to frame the site.
// See support table here: https://caniuse.com/mdn-http_headers_csp_content-security-policy_frame-ancestors.
func Allow(reason string, reportOnly bool, hostnames ...string) internalunsafeframing.AllowList {
	if reason == "" {
		panic("reason cannot be empty")
	}
	return internalunsafeframing.AllowList{
		ReportOnly: reportOnly,
		Hostnames:  hostnames,
	}
}
