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

// Package xsrf provides a safehttp.Interceptor that ensures Cross-Site Request
// Forgery protection by verifying the incoming requests, rejecting those
// requests that are suspected to be part of an attack.
package angularxsrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/go-safeweb/safehttp"
)

var statePreservingMethods = map[string]bool{
	safehttp.MethodGet:     true,
	safehttp.MethodHead:    true,
	safehttp.MethodOptions: true,
}

// Interceptor provides protection against Cross-Site Request Forgery attacks
// for Angular's XHR requests.
//
// See https://docs.angularjs.org/api/ng/service/$http#cross-site-request-forgery-xsrf-protection for more details.
type Interceptor struct {
	// TokenCookieName is the name of the seesion cookie that holds the XSRF
	// token. In order to prevent collisions when multiple applications share
	// the same domain or subdomain, each application should set a unique name
	// for the cookie.
	TokenCookieName string
	// TokenHeaderName is the name of the HTTP header that also holds the XSRF
	// token.
	TokenHeaderName string
}

// Before should be executed before directing the safehttp.IncomingRequest to
// the handler to ensure it is not part of a Cross-Site Request Forgery attacks.
//
// It will check for the presence of a matching XSRF token, generated on the
// first page access, in both a cookie and a header. Their names should be set
// when the Interceptor is created.
func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	c, err := r.Cookie(it.TokenCookieName)
	if err != nil {
		if !statePreservingMethods[r.Method()] {
			return w.WriteError(safehttp.StatusForbidden)
		}
		return safehttp.NotWritten()
	}

	tok := r.Header.Get(it.TokenHeaderName)
	if tok == "" || tok != c.Value() {
		return w.WriteError(safehttp.StatusUnauthorized)
	}

	return safehttp.NotWritten()
}

func (it *Interceptor) addAngularTokenCookie(w *safehttp.ResponseWriter) error {
	tok := make([]byte, 20)
	if _, err := rand.Read(tok); err != nil {
		return fmt.Errorf("crypto/rand.Read: %v", err)
	}
	c := safehttp.NewCookie(it.TokenCookieName, base64.StdEncoding.EncodeToString(tok))

	c.SetSameSite(safehttp.SameSiteStrictMode)
	c.SetPath("/")
	// Set the duration of the token cookie to 24 hours.
	c.SetMaxAge(86400)
	// Needed in order to make the cookie accessible by JavaScript
	// running on the user's domain.
	c.DisableHTTPOnly()

	if err := w.SetCookie(c); err != nil {
		return err
	}
	return nil
}

// Commit generates a cryptographically random cookie on the first state
// preserving request (GET, HEAD or OPTION) and sets it in the response. On
// every subsequent request the cookie is expected alongside a header that
// matches its value.
func (it *Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {

	_, err := r.Cookie(it.TokenCookieName)
	if err == nil {
		return safehttp.NotWritten()
	}

	if !statePreservingMethods[r.Method()] {
		return w.WriteError(safehttp.StatusForbidden)
	}

	err = it.addAngularTokenCookie(w)
	if err != nil {
		// A 500 error is returned when the plugin fails to set the Set-Cookie
		// header in the response writer as this is a server misconfiguration.
		return w.WriteError(safehttp.StatusInternalServerError)
	}
	return safehttp.NotWritten()
}

// OnError is a no-op, required to satisfy the safehttp.Interceptor interface.
func (it *Interceptor) OnError(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {
	return safehttp.NotWritten()
}
