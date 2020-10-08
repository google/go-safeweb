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
package xsrf

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/go-safeweb/safehttp"
	"golang.org/x/net/xsrftoken"
)

const (
	// TokenKey is the form key used when sending the token as part of POST
	// request.
	TokenKey    = "xsrf-token"
	cookieIDKey = "xsrf-cookie"

	// Angular's XHR XSRF token support.
	angularTokenCookie = "XSRF-TOKEN"
	angularTokenHeader = "X-XSRF-TOKEN"
)

var statePreservingMethods = map[string]bool{
	safehttp.MethodGet:     true,
	safehttp.MethodHead:    true,
	safehttp.MethodOptions: true,
}

// Interceptor implements XSRF protection.
type Interceptor struct {
	// SecretAppKey uniquely identifies each registered service and should have
	// high entropy as it is used for generating the XSRF token.
	SecretAppKey string
	// AngularProtection is used to check whether the interceptor should provide
	// protection against XSRF for Angular's XHR requests.
	//
	// See https://docs.angularjs.org/api/ng/service/$http#cross-site-request-forgery-xsrf-protection for more details.
	AngularProtection bool
}

var _ safehttp.Interceptor = &Interceptor{}

type tokenCtxKey struct{}

// Token extracts the XSRF token from the incoming request. If it is not
// present, it returns a non-nil error.
func Token(r *safehttp.IncomingRequest) (string, error) {
	tok := r.Context().Value(tokenCtxKey{})
	if tok == nil {
		return "", errors.New("xsrf token not found")
	}
	return tok.(string), nil
}

func addCookieID(w *safehttp.ResponseWriter) (*safehttp.Cookie, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("crypto/rand.Read: %v", err)
	}

	c := safehttp.NewCookie(cookieIDKey, base64.StdEncoding.EncodeToString(buf))
	c.SetSameSite(safehttp.SameSiteStrictMode)
	if err := w.SetCookie(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (it *Interceptor) addAngularTokenCookie(w *safehttp.ResponseWriter) error {
	tok := make([]byte, 20)
	if _, err := rand.Read(tok); err != nil {
		return fmt.Errorf("crypto/rand.Read: %v", err)
	}
	c := safehttp.NewCookie(angularTokenCookie, base64.StdEncoding.EncodeToString(tok))

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

// Before should be executed before directing the safehttp.IncomingRequest to
// the handler to ensure it is not part of a Cross-Site Request
// Forgery attack.
//
// If Angular protection is not set, on first user visit through a state
// preserving request (GET, HEAD  or OPTIONS), a nonce-based cookie will be set
// in the response as a way to distinguish between users and prevent pre-login
// XSRF attacks. The cookie will  be used in the token generation and
// verification algorithm and is expected to be present in all subsequent incoming requests.
//
// For every authorized request, the interceptor will also generate a
// cryptographically-safe XSRF token using the appKey, the cookie and the path
// visited. This can be later extracted using Token and should be injected as a
// hidden input field in HTML forms.
//
// In case of state changing requests (all except GET, HEAD and OPTIONS), the
// interceptor checks for the presence of the XSRF token in the request body
// (expected to have been injected) and validates it.
//
// If Angular protection is enabled, a token will be generated on the first GET
// request and assigned to the  the XSRF-TOKEN cookie, which will be set in the
// response. On every subsequent request, the cookie is expected alongside the
// X-XSRF-TOKEN header and their values should match. In case they don't, the
// request will be rejected.
func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if it.AngularProtection {
		c, err := r.Cookie(angularTokenCookie)
		if err != nil {
			if r.Method() != safehttp.MethodGet {
				return w.WriteError(safehttp.StatusUnauthorized)
			}
			err := it.addAngularTokenCookie(w)
			if err != nil {
				// An error is returned when the plugin fails to Set the Set-Cookie header in the response writer as this is a server misconfiguration
				return w.WriteError(safehttp.StatusInternalServerError)
			}
			return safehttp.NotWritten()
		}

		tok := r.Header.Get(angularTokenHeader)
		if tok == "" {
			return w.WriteError(safehttp.StatusUnauthorized)
		}

		if tok != c.Value() {
			return w.WriteError(safehttp.StatusForbidden)
		}
		return safehttp.NotWritten()

	}

	m := r.Method()
	needsValidation := !statePreservingMethods[m]
	cookieID, err := r.Cookie(cookieIDKey)
	if err != nil {
		if needsValidation {
			return w.WriteError(safehttp.StatusForbidden)
		}
		cookieID, err = addCookieID(w)
		if err != nil {
			// An error is returned when the plugin fails to Set the Set-Cookie
			// header in the response writer as this is a server misconfiguration.
			return w.WriteError(safehttp.StatusInternalServerError)
		}
	}

	actionID := r.URL.Path()
	if needsValidation {
		f, err := r.PostForm()
		if err != nil {
			// We fallback to checking whether the form is multipart. Both types
			// are valid in an incoming request as long as the XSRF token is
			// present.
			mf, err := r.MultipartForm(32 << 20)
			if err != nil {
				return w.WriteError(safehttp.StatusBadRequest)
			}
			f = &mf.Form
		}

		tok := f.String(TokenKey, "")
		if f.Err() != nil || tok == "" {
			return w.WriteError(safehttp.StatusUnauthorized)
		}

		if !xsrftoken.Valid(tok, it.SecretAppKey, cookieID.Value(), actionID) {
			return w.WriteError(safehttp.StatusForbidden)
		}
	}

	tok := xsrftoken.Generate(it.SecretAppKey, cookieID.Value(), actionID)
	r.SetContext(context.WithValue(r.Context(), tokenCtxKey{}, tok))
	return safehttp.NotWritten()
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface,
// for  Angular protection. Otherwise, it adds the XSRF token corresponding to
// the safehttp.TemplateResponse  with key "XSRFToken". The token corresponds to
// the  user information found in the request.
func (it *Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {
	if it.AngularProtection {
		return safehttp.NotWritten()
	}

	tmplResp, ok := resp.(safehttp.TemplateResponse)
	if !ok {
		return safehttp.NotWritten()
	}

	tok, err := Token(r)
	if err != nil {
		// The token should have been added in the Before stage and if that is
		// not the case, a server misconfiguration occured.
		return w.WriteError(safehttp.StatusInternalServerError)
	}

	// TODO(maramihali@): Change the key when function names are exported by
	// htmlinject
	// TODO: what should happen if the XSRFToken key is not present in the
	// tr.FuncMap?
	tmplResp.FuncMap["XSRFToken"] = func() string { return tok }
	return safehttp.NotWritten()
}
