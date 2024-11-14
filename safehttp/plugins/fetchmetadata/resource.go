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

package fetchmetadata

import (
	"log"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata/internalunsafefetchmetadata"
)

var (
	navigationalDests = map[string]bool{
		"document":        true,
		"nested-document": true,
	}
	statePreservingMethods = map[string]bool{
		safehttp.MethodGet:  true,
		safehttp.MethodHead: true,
	}
)

// ResourceIsolationPolicy protects resources.
//
// See https://web.dev/fetch-metadata/ for more details.
func ResourceIsolationPolicy() *Policy {
	return &Policy{
		isAllowed: func(r *safehttp.IncomingRequest) bool {
			h := r.Header
			if h.Get("Sec-Fetch-Site") != "cross-site" {
				// The request is allowed to pass because one of the following applies:
				// - Fetch Metadata is not supported by the browser
				// - the request is same-origin, same-site or caused by the user
				// explicitly interacting with the user-agent
				return true
			}

			method := r.Method()
			mode := h.Get("Sec-Fetch-Mode")
			dest := h.Get("Sec-Fetch-Dest")
			// Allow CORS options requests if neither Mode nor Dest is set.
			// https://github.com/w3c/webappsec-fetch-metadata/issues/35
			// https://bugs.chromium.org/p/chromium/issues/detail?id=979946
			if mode == "" && dest == "" && method == safehttp.MethodOptions {
				return true
			}

			if navigationalModes[mode] && navigationalDests[dest] && statePreservingMethods[method] {
				// The request is cross-site, but a simple top-level navigation from a
				// safe destination so we allow it to pass.
				return true
			}
			if safehttp.IsLocalDev() {
				log.Println("fetchmetadata plugin resource protection detected a potentially malicious request")
			}
			return false
		},
		match: func(cfg safehttp.InterceptorConfig) bool {
			_, ok := cfg.(internalunsafefetchmetadata.DisableResourceIsolationPolicy)
			return ok
		},
		skip: func(cfg safehttp.InterceptorConfig) (skip, skipReports bool) {
			if override, ok := cfg.(internalunsafefetchmetadata.DisableResourceIsolationPolicy); ok {
				return true, override.SkipReports
			}
			return false, false
		},
		signal: "RESOURCE_ISOLATION",
	}
}
