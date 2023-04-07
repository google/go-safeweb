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

package safehttp

import (
	"net/http"
)

// A Cookie represents an HTTP cookie as sent in the Set-Cookie header of an
// HTTP response or the Cookie header of an HTTP request.
//
// See https://tools.ietf.org/html/rfc6265 for details.
type Cookie struct {
	wrapped *http.Cookie
}

// NewCookie creates a new Cookie with safe default settings.
// Those safe defaults are:
//   - Secure: true (if the framework is not in dev mode)
//   - HttpOnly: true
//   - SameSite: Lax
//
// For more info about all the options, see:
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
func NewCookie(name, value string) *Cookie {
	devMu.RLock()
	defer devMu.RUnlock()
	return &Cookie{
		&http.Cookie{
			Name:     name,
			Value:    value,
			Secure:   !isLocalDev,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}
}

// SameSite allows a server to define a cookie attribute making it impossible for
// the browser to send this cookie along with cross-site requests. The main
// goal is to mitigate the risk of cross-origin information leakage, and provide
// some protection against cross-site request forgery attacks.
//
// See https://tools.ietf.org/html/draft-ietf-httpbis-cookie-same-site-00 for details.
type SameSite int

const (
	// SameSiteLaxMode allows sending cookies with same-site requests and
	// cross-site top-level navigations.
	SameSiteLaxMode SameSite = iota + 1
	// SameSiteStrictMode allows sending cookie only with same-site requests.
	SameSiteStrictMode
	// SameSiteNoneMode allows sending cookies with all requests, including the
	// ones made cross-origin.
	SameSiteNoneMode
)

// SameSite sets the SameSite attribute.
func (c *Cookie) SameSite(s SameSite) {
	switch s {
	case SameSiteLaxMode:
		c.wrapped.SameSite = http.SameSiteLaxMode
	case SameSiteStrictMode:
		c.wrapped.SameSite = http.SameSiteStrictMode
	case SameSiteNoneMode:
		c.wrapped.SameSite = http.SameSiteNoneMode
	}
}

// SetMaxAge sets the MaxAge attribute.
//
// - MaxAge = 0 means no 'Max-Age' attribute specified.
// - MaxAge < 0 means delete cookie now, equivalently 'Max-Age: 0'
// - MaxAge > 0 means Max-Age attribute present and given in seconds
func (c *Cookie) SetMaxAge(maxAge int) {
	c.wrapped.MaxAge = maxAge
}

// Path sets the path attribute.
func (c *Cookie) Path(path string) {
	c.wrapped.Path = path
}

// Domain sets the domain attribute.
func (c *Cookie) Domain(domain string) {
	c.wrapped.Domain = domain
}

// DisableSecure disables the secure attribute.
func (c *Cookie) DisableSecure() {
	c.wrapped.Secure = false
}

// DisableHTTPOnly disables the HttpOnly attribute.
func (c *Cookie) DisableHTTPOnly() {
	c.wrapped.HttpOnly = false
}

// Name returns the name of the cookie.
func (c *Cookie) Name() string {
	return c.wrapped.Name
}

// Value returns the value of the cookie.
func (c *Cookie) Value() string {
	return c.wrapped.Value
}

// String returns the serialization of the cookie for use in a Set-Cookie
// response header. If c is nil or c.Name() is invalid, the empty string is
// returned.
func (c *Cookie) String() string {
	return c.wrapped.String()
}
