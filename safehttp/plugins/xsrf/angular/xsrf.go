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

package xsrfangular

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"time"
)

// Interceptor provides protection against Cross-Site Request Forgery attacks
// for Angular's XHR requests.
//
// See https://docs.angularjs.org/api/ng/service/$http#cross-site-request-forgery-xsrf-protection for more details.
type Interceptor struct {
	// TokenCookieName is the name of the session cookie that holds the XSRF
	// token.
	TokenCookieName string
	// TokenHeaderName is the name of the HTTP header that holds the XSRF token.
	TokenHeaderName string
}

var _ safehttp.Interceptor = &Interceptor{}

// Default creates an Interceptor with TokenCookieName set to XSRF-TOKEN and
// TokenHeaderName set to X-XSRF-TOKEN, their default values. However, in order
// to prevent collisions when multiple applications share the same domain or
// subdomain, each application should set a unique name for the cookie.
//
// See https://docs.angularjs.org/api/ng/service/$http#cross-site-request-forgery-xsrf-protection for more details.
func Default() *Interceptor {
	return &Interceptor{
		TokenCookieName: "XSRF-TOKEN",
		TokenHeaderName: "X-XSRF-TOKEN",
	}
}

// Before checks for the presence of a matching XSRF token, generated on the
// first page access, in both a cookie and a header. Their names should be set
// when the Interceptor is created.
func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if xsrf.StatePreserving(r) {
		return safehttp.NotWritten()
	}

	c, err := r.Cookie(it.TokenCookieName)
	if err != nil || c.Value() == "" {
		return w.WriteError(safehttp.StatusForbidden)
	}

	tok := r.Header.Get(it.TokenHeaderName)
	if tok == "" || tok != c.Value() {
		// JavaScript has access only to cookies from the domain it's running
		// on. Hence, if the same token is found in both the cookie and the
		// header, the request can be trusted.
		return w.WriteError(safehttp.StatusUnauthorized)
	}

	return safehttp.NotWritten()
}

func (it *Interceptor) addTokenCookie(w *safehttp.ResponseWriter) error {
	tok := make([]byte, 20)
	if _, err := rand.Read(tok); err != nil {
		return fmt.Errorf("crypto/rand.Read: %v", err)
	}
	c := safehttp.NewCookie(it.TokenCookieName, base64.StdEncoding.EncodeToString(tok))

	c.SetSameSite(safehttp.SameSiteStrictMode)
	c.SetPath("/")
	day := 24 * time.Hour
	c.SetMaxAge(int(day.Seconds()))
	// Needed in order to make the cookie accessible by JavaScript
	// running on the same domain.
	c.DisableHTTPOnly()

	return w.SetCookie(c)
}

// Commit generates a cryptographically secure random cookie on the first state
// preserving request (GET, HEAD or OPTION) and sets it in the response. On
// every subsequent request the cookie is expected alongside a header that
// matches its value.
func (it *Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {
	if c, err := r.Cookie(it.TokenCookieName); err == nil && c.Value() != "" {
		// The XSRF cookie is there so we don't need to do anything else.
		return safehttp.NotWritten()
	}

	if !xsrf.StatePreserving(r) {
		// This should never happen as, if this is a state-changing request and
		// it lacks the cookie, it would've been already rejected by Before.
		return w.WriteError(safehttp.StatusInternalServerError)
	}

	err := it.addTokenCookie(w)
	if err != nil {
		// This is a server misconfiguration.
		return w.WriteError(safehttp.StatusInternalServerError)
	}
	return safehttp.NotWritten()
}

// OnError is a no-op, required to satisfy the safehttp.Interceptor interface.
func (it *Interceptor) OnError(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {
	return safehttp.NotWritten()
}
