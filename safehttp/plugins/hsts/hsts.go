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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-safeweb/safehttp"
)

// Plugin implements automatic HSTS functionality.
// See https://tools.ietf.org/html/rfc6797 for more info.
type Plugin struct {
	// MaxAge is the duration that the browser should remember
	// that a site is only to be accessed using HTTPS. MaxAge
	// must be positive.
	MaxAge time.Duration

	// DisableIncludeSubDomains disables the includeSubDomains directive.
	// When DisableIncludeSubDomains is false, all subdomains
	// of the domain where this service is hosted will also be added
	// to the browsers HSTS list.
	DisableIncludeSubDomains bool

	// Preload enables the preload directive.
	// This should only be enabled if this site should be
	// added to the browser HSTS preload list, which is supported
	// by all major browsers. See https://hstspreload.org/ for
	// more info.
	Preload bool

	// BehindProxy controls how the plugin should behave with regards
	// to HTTPS. If this server is behind a proxy that terminates HTTPS
	// traffic then this should be enabled. If this is enabled
	// then the plugin will always send the Strict-Transport-Security
	// header and will not redirect HTTP traffic to HTTPS traffic.
	BehindProxy bool
}

// NewPlugin creates a new HSTS plugin with safe defaults.
// These safe defaults are:
//  - max-age set to two years.
//  - includeSubDomains is enabled.
//  - preload is disabled.
func NewPlugin() Plugin {
	return Plugin{MaxAge: 63072000 * time.Second} // two years in seconds
}

// Before should be executed before the request is sent to the handler.
// The function redirects HTTP requests to HTTPS. When HTTPS traffic
// is received the Strict-Transport-Security header is applied to the
// response.
func (p *Plugin) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	if p.MaxAge < 0 {
		return w.ServerError(safehttp.StatusInternalServerError, "Internal Server Error")
	}

	if !p.BehindProxy && r.TLS == nil {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			return w.ServerError(safehttp.StatusInternalServerError, "Internal Server Error")
		}
		u.Scheme = "https"
		return w.Redirect(r, u.String(), safehttp.StatusMovedPermanently)
	}

	var value strings.Builder
	value.WriteString("max-age=")
	value.WriteString(strconv.FormatInt(int64(p.MaxAge.Seconds()), 10))
	if !p.DisableIncludeSubDomains {
		value.WriteString("; includeSubDomains")
	}
	if p.Preload {
		value.WriteString("; preload")
	}
	h := w.Header()
	if err := h.Set("Strict-Transport-Security", value.String()); err != nil {
		// TODO(@mattiasgrenfeldt): Replace the response with an actual saferesponse somehow.
		return w.ServerError(safehttp.StatusInternalServerError, "Internal Server Error")
	}
	if _, err := h.Claim("Strict-Transport-Security"); err != nil {
		return w.ServerError(safehttp.StatusInternalServerError, "Internal Server Error")
	}
	return safehttp.Result{}
}
