// Copyright 2020 Google LLC
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

// Package coop provides Cross-Origin-Opener-Policy protection. Specification: https://html.spec.whatwg.org/#cross-origin-opener-policies
package coop

import (
	"github.com/google/go-safeweb/safehttp"
)

var _ safehttp.Interceptor = Interceptor{}

// Mode represents a COOP mode.
type Mode string

const (
	// SameOrigin is the strictest and safest COOP available: windows can keep a reference of windows they open only if they are same-origin.
	SameOrigin Mode = "same-origin"
	// SameOriginAllowPopups relaxes the same-origin COOP: windows on this origin that open other windows are allowed to keep a reference, but the opposite is not valid.
	SameOriginAllowPopups Mode = "same-origin-allow-popups"
	// UnsafeNone disables COOP: this is the default value in browsers.
	UnsafeNone Mode = "unsafe-none"
)

// Policy represents a Cross-Origin-Opener-Policy value.
type Policy struct {
	// Mode is the mode for the policy.
	Mode Mode
	// ReportingGroup is an optional reporting group that needs to be defined with the Reporting API.
	ReportingGroup string
	// ReportOnly makes the policy report-only if set.
	ReportOnly bool
}

// String serializes the policy. The returned value can be used as a header value.
func (p Policy) String() string {
	if p.ReportingGroup == "" {
		return string(p.Mode)
	}
	return string(p.Mode) + `; report-to "` + p.ReportingGroup + `"`
}

type serializedPolicies struct {
	rep []string
	enf []string
}

func serializePolicies(policies ...Policy) serializedPolicies {
	var s serializedPolicies
	for _, p := range policies {
		if p.ReportOnly {
			s.rep = append(s.rep, p.String())
		} else {
			s.enf = append(s.enf, p.String())
		}
	}
	return s
}

// NewInterceptor constructs an interceptor that applies the given policies.
func NewInterceptor(policies ...Policy) Interceptor {
	return Interceptor(serializePolicies(policies...))
}

// Default returns a same-origin enforcing interceptor with the given (potentially empty) report group.
func Default(reportGroup string) Interceptor {
	return NewInterceptor(Policy{Mode: SameOrigin, ReportingGroup: reportGroup})
}

// Interceptor is the interceptor for COOP.
type Interceptor serializedPolicies

// Before claims and sets the Report-Only and Enforcement headers for COOP.
func (it Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	if cfg != nil {
		// We got an override, run its Before phase instead.
		return Interceptor(cfg.(Overrider)).Before(w, r, nil)
	}
	w.Header().Claim("Cross-Origin-Opener-Policy")(it.enf)
	w.Header().Claim("Cross-Origin-Opener-Policy-Report-Only")(it.rep)
	return safehttp.NotWritten()
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (it Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}

// Match recognizes Overriders as COOP configurations.
func (it Interceptor) Match(cfg safehttp.InterceptorConfig) bool {
	_, ok := cfg.(Overrider)
	return ok
}

// Overrider is a safehttp.InterceptorConfig that allows to override COOP for a specific handler.
type Overrider serializedPolicies

// Override creates an Overrider with the given policies.
func Override(reason string, policies ...Policy) Overrider {
	return Overrider(serializePolicies(policies...))
}
