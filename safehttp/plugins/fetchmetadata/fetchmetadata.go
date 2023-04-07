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

// Package fetchmetadata provides Fetch-Metadata based protections.
package fetchmetadata

import (
	"log"

	"github.com/google/go-safeweb/safehttp"
)

var (
	navigationalModes = map[string]bool{
		"navigate":        true,
		"nested-navigate": true,
	}
)

// TODO(empijei): implement NIP as soon as it's production ready.

// Policy is a security policy based on Fetch Metadata.
//
// See https://web.dev/fetch-metadata/ for more.
type Policy struct {
	isAllowed  func(*safehttp.IncomingRequest) bool
	match      func(safehttp.InterceptorConfig) bool
	skip       func(cfg safehttp.InterceptorConfig) (skip, skipReports bool)
	signal     string
	navigate   *safehttp.URL
	ReportOnly bool
}

// Before implements the Fetch Metadata validation and signals logic.
func (p *Policy) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	skip, skipReports := p.skip(cfg)
	if p.ReportOnly {
		skip = true
	}

	if p.isAllowed(r) {
		return safehttp.NotWritten()
	}

	if !skipReports {
		// TODO(empijei): report this.
		log.Printf("Request for %s %q should be blocked by %s. Actually_blocked=%v", r.Method(), r.URL().String(), p.signal, !skip)
	}
	if skip {
		return safehttp.NotWritten()
	}
	if safehttp.IsLocalDev() {
		log.Println("fetchmetadata plugin blocked a potentially malicious request")
	}
	if p.navigate != nil {
		return safehttp.Redirect(w, r, p.navigate.String(), safehttp.StatusSeeOther)
	}
	return w.WriteError(safehttp.StatusForbidden)
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (p *Policy) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}

// Match recongnizes configs to disable fetch metadata protection.
func (p *Policy) Match(cfg safehttp.InterceptorConfig) bool { return p.match(cfg) }
