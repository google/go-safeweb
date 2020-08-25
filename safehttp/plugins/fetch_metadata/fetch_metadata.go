// Copyright 2020 Google LLC
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

package fetchmetadata

import (
	"github.com/google/go-safeweb/safehttp"
)

// LoggingService is a user-provided service for logging Fetch Metadata policy
// violations
type LoggingService interface {
	Log(ir *safehttp.IncomingRequest)
}

// Plugin implements Fetch Metadata functionality.
// See https://www.w3.org/TR/fetch-metadata/ for more details.
type Plugin struct {
	// Policy is the Fetch Metadata policy that decides whether a request should
	// be allowed to pass or be rejected. This will be set to a default Resource
	// Isolation Policy, but can be changed by the user.
	//
	// See https://web.dev/fetch-metadata/ for more information on the default policy
	Policy func(*safehttp.IncomingRequest, map[string]bool) bool
	// NavIsolation indicates whether the Navigation Isolation Policy should
	// also be applied to the request as an additional layer of checking. This
	// provides a way to mitigate clickjacking and reflected XSS by rejecting
	// all cross-site navigations unless coming from origins where CORS is
	// explicitly enabled. This can be used independently from a Resource
	// Isolation Policy.
	//
	// WARNING: This is still an experimental feature and will be disabled by default.
	NavIsolation bool
	reportOnly   bool
	logger       LoggingService
	allowedCORS  map[string]bool
}

// NewPlugin creates a new Fetch Metadata plugin in enforce mode using the
// defaultPolicy and a set of user-provided origins on which CORS is enabled.
func NewPlugin(origins ...string) *Plugin {
	m := map[string]bool{}
	for _, origin := range origins {
		m[origin] = true
	}
	return &Plugin{
		allowedCORS: m,
		Policy:      defaultPolicy,
	}
}

func defaultPolicy(r *safehttp.IncomingRequest, allowedCORS map[string]bool) bool {
	h := r.Header
	switch h.Get("Sec-Fetch-Site") {
	case "":
		// Fetch Metadata is not supported by the browser so we allow the
		// request to pass.
		return true
	case "same-origin", "same-site", "none":
		// The request originated from a site served by your own server
		// ("same-origin"), a subdomain of your site ("same-site", e.g.
		// bar.foo.com made a request to foo.com) or was caused by the user
		// explicitly interacting with the user-agent ("none"). Therefore it is
		// allowed to pass.
		return true
	}

	if m := r.Method(); h.Get("Sec-Fetch-Mode") == "navigate" && (m == safehttp.MethodGet || m == safehttp.MethodHead) {
		if dest := h.Get("Sec-Fetch-Dest"); dest == "object" || dest == "embed" {
			// The request is cross-site and originates from <object> or <embed>
			// so it is rejected.
			return false
		}
		// The request is cross-site, but a simple top-level navigation so we
		// allow it to pass.
		return true
	}

	origin := h.Get("Origin")
	if _, ok := allowedCORS[origin]; ok && h.Get("Sec-Fetch-Mode") == "cors" {
		// The request is cross-site but sent from an
		// origin where Cross-Origin Resource Sharing is explicitly allowed.
		return true
	}

	// The request is cross-site, not navigational and sent from an endpoint
	// where CORS is not enabled therefore it is rejected.
	return false
}

func navigationIsolationPolicy(r *safehttp.IncomingRequest, allowedCORS map[string]bool) bool {
	h := r.Header
	origin := h.Get("Origin")
	if h.Get("Sec-Fetch-Site") == "cross-site" && h.Get("Sec-Fetch-Mode") == "navigate" {
		if _, ok := allowedCORS[origin]; !ok {
			return false
		}
	}
	return true
}

// SetReportMode sets the Fetch Metadata policy mode to "report". This will
// allow requests that violate the policy to pass, but will report the violation
// using the LoggingService. The method will panic if no LoggingService is
// provided.
func (p *Plugin) SetReportMode(logger LoggingService) {
	if logger == nil {
		panic("logging service required for Fetch Metadata report mode")
	}
	p.logger = logger
	p.reportOnly = true
}

// SetEnforceMode sets the Fetch Metadata policy mode to "enforce". This will
// reject any requests that violate the policy provided by the plugin.
func (p *Plugin) SetEnforceMode() {
	p.reportOnly = false
}

// Before validates the safehttp.IncomingRequest using the Fetch Metadata policy
// provided by the  plugin. It only allows request to pass if they conform to
// the policy, if it's sent from a CORS endpoint, or if the mode is set to
// "report", in which case the request is allowed to pass but the violation is
// reported. Moreover, if the Navigation Isolation Policy is enabled, the
// request will also be validated against it.  If the browser does not have
// Fetch Metadata support implemented, the policy will not be applied and all
// requests will be allowed to pass.
func (p *Plugin) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	if !p.Policy(r, p.allowedCORS) {
		if !p.reportOnly {
			return w.ClientError(safehttp.StatusForbidden)
		}
		p.logger.Log(r)
		return safehttp.Result{}
	}
	if p.NavIsolation && !navigationIsolationPolicy(r, p.allowedCORS) {
		if !p.reportOnly {
			return w.ClientError(safehttp.StatusForbidden)
		}
		p.logger.Log(r)
		return safehttp.Result{}
	}
	return safehttp.Result{}
}
