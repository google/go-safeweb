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

// Package unsaferesourcepolicy can be used to disable Fetch Metadata protections on specific handler registration.
//
// This package is visibility restricted and its uses require a review by the security team.
package unsaferesourcepolicy

import (
	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata/internalunsafefetchmetadata"
)

// Disable switches the resource isolation policy to report-only.
// If skipReports is true, it will completely disable it.
func Disable(reason string, skipReports bool) internalunsafefetchmetadata.DisableResourceIsolationPolicy {
	if reason == "" {
		panic("reason cannot be empty")
	}
	return internalunsafefetchmetadata.DisableResourceIsolationPolicy{SkipReports: skipReports}
}
