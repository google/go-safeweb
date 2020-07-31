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

package hsts

import (
	"strconv"
	"strings"

	"github.com/google/go-safeweb/safehttp"
)

// Plugin implements automatic HSTS functionality.
type Plugin struct {
	maxAge            uint64
	includeSubDomains bool
	preload           bool
}

// NewPlugin creates a new HSTS plugin with safe defaults.
func NewPlugin() Plugin {
	return Plugin{
		maxAge:            63072000, // two years in seconds
		includeSubDomains: true,
		preload:           false,
	}
}

// Before should be executed before the request is sent to the handler.
// The function redirects HTTP requests to HTTPS. When HTTPS traffic
// is received the Strict-Transport-Security header is applied to the
// response.
func (p *Plugin) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	if r.TLS == nil {
		r.URL.Scheme = "https"
		return w.Redirect(r, r.URL.String(), 301)
	}

	value := strings.Builder{}
	value.WriteString("max-age=" + strconv.FormatUint(p.maxAge, 10))
	if p.includeSubDomains {
		value.WriteString("; includeSubDomains")
	}
	if p.preload {
		value.WriteString("; preload")
	}
	h := w.Header()
	if err := h.Set("Strict-Transport-Security", value.String()); err != nil {
		// TODO(@mattiasgrenfeldt): Replace the response with an actual saferesponse somehow.
		return w.ServerError(500, "Internal Server Error")
	}
	return safehttp.Result{}
}

// EnablePreload enables the preload directive.
// This should only be enabled if this site should be
// added to the browser HSTS preload list which is supported
// by all major browsers. See https://hstspreload.org/ for
// more info.
func (p *Plugin) EnablePreload() {
	p.preload = true
}

// DisableIncludeSubDomains disables the includeSubDomains
// directive. When includeSubDomains is enabled, all subdomains
// of the domain where this service is hosted will also be added
// to the browsers HSTS list. This method disables this feature.
func (p *Plugin) DisableIncludeSubDomains() {
	p.includeSubDomains = false
}
