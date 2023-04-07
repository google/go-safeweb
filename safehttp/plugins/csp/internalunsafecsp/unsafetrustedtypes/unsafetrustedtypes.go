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

// Package  unsafetrustedtypes can be used to disable Trusted Types protections on specific handler registration.
//
// Usage of this package should require a security review.
package unsafetrustedtypes

import "github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp"

// Disable switches the trusted types policy to report-only.
// If skipReports is true, it will completely disable it.
func Disable(reason string, skipReports bool) internalunsafecsp.DisableTrustedTypes {
	if reason == "" {
		panic("reason cannot be empty")
	}
	return internalunsafecsp.DisableTrustedTypes{SkipReports: skipReports}
}
