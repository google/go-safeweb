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

// Package unsafestrictcsp can be used to disable Strict CSP protections on specific handler registration.
//
// Usage of this package should require a security review.
package unsafestrictcsp

import "github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp"

// Disable switches Strict CSP to report-only.
// If skipReports is true, it will completely disable it.
func Disable(reason string, skipReports bool) internalunsafecsp.DisableStrict {
	if reason == "" {
		panic("reason cannot be empty")
	}
	return internalunsafecsp.DisableStrict{SkipReports: skipReports}
}
