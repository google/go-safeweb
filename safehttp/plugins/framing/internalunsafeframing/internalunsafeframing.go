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

// Package internalunsafeframing is used internally to override Framing protections.
package internalunsafeframing

// Disable turns framing protection to report-only where possible.
type Disable struct {
	// SkipReports completely disables framing protectcion.
	SkipReports bool
}

// AllowList selectively allows framing.
//
// Please note that on older browsers this is equivalent to Disable.
type AllowList struct {
	// ReportOnly sets the policy to Report-Only instead of enforcing.
	ReportOnly bool
	// Hostnames is a list of origins (with potential wildcards) that will be able to frame the site.
	// Wildcards must follow the CSP specification:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-ancestors.
	Hostnames []string
}
