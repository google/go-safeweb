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

package fetchmetadata

import (
	"log"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"
)

var (
	frameableDests = map[string]bool{
		"frame":  true,
		"iframe": true,
		"embed":  true,
		"object": true,
	}
)

// FramingIsolationPolicy protects from framing attacks.
//
// See https://xsleaks.dev/docs/defenses/isolation-policies/framing-isolation/#implementation-with-fetch-metadata
func FramingIsolationPolicy() *Policy {
	return &Policy{
		isAllowed: func(r *safehttp.IncomingRequest) bool {
			h := r.Header
			mode := h.Get("Sec-Fetch-Mode")
			dest := h.Get("Sec-Fetch-Dest")
			site := h.Get("Sec-Fetch-Site")
			if mode == "" || dest == "" || site == "" {
				return true
			}
			if !navigationalModes[mode] {
				// Allow non-navigational requests.
				return true
			}
			if !frameableDests[dest] {
				// Allow non-frameable requests.
				return true
			}
			if site == "same-origin" || site == "none" {
				return true
			}
			if safehttp.IsLocalDev() {
				log.Println("fetchmetadata plugin framing protection detected a potentially malicious request")
			}
			return false
		},
		match: func(cfg safehttp.InterceptorConfig) bool {
			switch cfg.(type) {
			case internalunsafeframing.Disable, internalunsafeframing.AllowList:
				return true
			}
			return false
		},
		skip: func(cfg safehttp.InterceptorConfig) (skip, skipReports bool) {
			switch c := cfg.(type) {
			case internalunsafeframing.Disable:
				return true, c.SkipReports
			case internalunsafeframing.AllowList:
				// It is intended to get someone to frame us.
				// Since Fetch Metadata is not fine-grained we cannot tell if
				// a violation is intended, so we should just disable this completely.
				return true, true
			}
			return false, false
		},
		signal: "FRAMING_ISOLATION",
	}
}
