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

// Package fetchmetadata provides a safehttp.Interceptor that applies Fetch
// Metadata policies to incoming requests in order to protect applications
// against cross-origin attacks.
package fetchmetadata

import (
	"github.com/google/go-safeweb/safehttp"
)

// RequestLogger is used for logging Fetch Metadata policy violations;
type RequestLogger interface {
	Log(ir *safehttp.IncomingRequest, navigationIsolation bool)
}

var (
	navigationalModes = map[string]bool{
		"navigate":        true,
		"nested-navigate": true,
	}
	navigationalDest = map[string]bool{
		"document":        true,
		"nested-document": true,
	}
	statePreservingMethods = map[string]bool{
		safehttp.MethodGet:  true,
		safehttp.MethodHead: true,
	}
)

// Interceptor implements Fetch Metadata functionality.
// A pointer to the zero value for Interceptor is safe and ready to use.
//
// See https://www.w3.org/TR/fetch-metadata/ and
// https://web.dev/fetch-metadata/  for more details.
type Interceptor struct {
	// NavIsolation indicates whether the Navigation Isolation Policy should
	// be applied to the request before the Resource Isolation Policy as an
	// additional layer of checking. This provides a way to mitigate
	// clickjacking and reflected XSS by rejecting all cross-site navigations
	// unless targeted to endpoints that are CORS-protected.
	//
	// WARNING: This is still an experimental feature and is disabled by
	// default.
	NavIsolation bool
	// RedirectURL can optionally indicate an URL to which the user can be
	// redirected in case the Navigation Isolation policy rejects the request.
	RedirectURL *safehttp.URL
	Logger      RequestLogger
	reportOnly  bool
}

var _ safehttp.Interceptor = &Interceptor{}

func (p *Interceptor) checkResourceIsolationPolicy(r *safehttp.IncomingRequest) bool {
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

	if navigationalModes[mode] && navigationalDest[dest] && statePreservingMethods[method] {
		// The request is cross-site, but a simple top-level navigation from a
		// safe destination so we  allow it to pass.
		return true
	}
	// The request is cross-site and not a simple navigation or from an unsafe
	// destination so it is rejected.
	return false
}

func (p *Interceptor) checkNavigationIsolationPolicy(r *safehttp.IncomingRequest) bool {
	if h := r.Header; p.NavIsolation && h.Get("Sec-Fetch-Site") == "cross-site" && navigationalModes[h.Get("Sec-Fetch-Mode")] {
		return false
	}
	return true
}

// SetReportOnly sets the Fetch Metadata policy mode to "report". This will
// allow requests that violate the policy to pass, but will report the violation
// using the Logger. The method will panic if Logger is nil.
func (p *Interceptor) SetReportOnly() {
	if p.Logger == nil {
		panic("logging service required for Fetch Metadata report mode")
	}
	p.reportOnly = true
}

// SetEnforce sets the Fetch Metadata policy mode to "enforce". This will reject
// any requests that violates the policy provided by the plugin.
func (p *Interceptor) SetEnforce() {
	p.reportOnly = false
}

// Before validates the safehttp.IncomingRequest using the Resource Isolation
// Policy and, if enabled, the Navigation Isolation Policy. It only allows
// requests to pass if they conform to the policy, if it's targeted to one of
// the CORS-protected endpoints, specified when creating the plugin, or if the
// mode is set to  "report", in which case the request is allowed to pass but
// the  violation is reported. If a redirectURL was provided and the Navigation
// Isolation Policy is enabled and fails, the IncomingRequest will be
// redirected to redirectURL.
func (p *Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	rejected := false
	nav := false

	if !p.checkNavigationIsolationPolicy(r) {
		rejected = true
		nav = true
	}

	if !p.checkResourceIsolationPolicy(r) {
		rejected = true
	}

	if !rejected {
		// Both Navigation Isolation and Resource Isolation are happy.
		return safehttp.NotWritten()
	}

	// Overridden to be disabled.
	if d, ok := cfg.(Disable); ok {
		if !d.SkipReporting && p.Logger != nil {
			p.Logger.Log(r, nav)
		}
		return safehttp.NotWritten()
	}

	if p.Logger != nil {
		p.Logger.Log(r, nav)
	}
	if p.reportOnly {
		return safehttp.NotWritten()
	}
	if nav && p.RedirectURL != nil {
		return safehttp.Redirect(w, r, p.RedirectURL.String(), safehttp.StatusSeeOther)
	}
	return w.WriteError(safehttp.StatusForbidden)
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (p *Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}

// Disable disables Fetch Metadata protection and switches it to report-only.
type Disable struct {
	// SkipReporting also disables reporting.
	SkipReporting bool
}

// Match matches this config to *Interceptor.
func (Disable) Match(i safehttp.Interceptor) bool {
	_, ok := i.(*Interceptor)
	return ok
}
