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

package xsrfhtml

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/htmlinject"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"golang.org/x/net/xsrftoken"
)

const (
	// TokenKey is the form key used when sending the token as part of POST
	// request.
	TokenKey    = "xsrf-token"
	cookieIDKey = "xsrf-cookie"
)

// Interceptor implements XSRF protection.
type Interceptor struct {
	// SecretAppKey uniquely identifies each registered service and should have
	// high entropy as it is used for generating the XSRF token.
	SecretAppKey string
}

var _ safehttp.Interceptor = &Interceptor{}

func addCookieID(w safehttp.ResponseHeadersWriter) (*safehttp.Cookie, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("crypto/rand.Read: %v", err)
	}

	c := safehttp.NewCookie(cookieIDKey, base64.StdEncoding.EncodeToString(buf))
	c.SameSite(safehttp.SameSiteStrictMode)
	if err := w.AddCookie(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Before checks for the presence of a XSRF token in the body of state changing
// requests (all except GET, HEAD and OPTIONS) and validates it.
func (it *Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if xsrf.StatePreserving(r) {
		return safehttp.NotWritten()
	}

	cookieID, err := r.Cookie(cookieIDKey)
	if err != nil {
		return w.WriteError(safehttp.StatusForbidden)
	}

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

	if ok := xsrftoken.Valid(tok, it.SecretAppKey, cookieID.Value(), r.URL().Host()); !ok {
		return w.WriteError(safehttp.StatusForbidden)
	}

	return safehttp.NotWritten()
}

// Commit adds XSRF protection in the response, so the interceptor can
// distinguish between subsequent requests coming from an authorized user and
// requests that are potentially part of a Cross-Site Request Forgery attack.
//
// On first user visit through a state preserving request (GET, HEAD or
// OPTIONS), a nonce-based cookie is set in the response as a way to
// distinguish between users and prevent pre-login XSRF attacks. The cookie is
// then used in the token generation and verification algorithm and is expected
// to be present in all subsequent incoming requests.
//
// For every authorized request, the interceptor also generates a
// cryptographically-safe XSRF token using the appKey, the cookie and the path
// visited. This is then injected as a hidden input field in HTML forms.
func (it *Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
	cookieID, err := r.Cookie(cookieIDKey)
	if err != nil {
		if !xsrf.StatePreserving(r) {
			// Not a state preserving request, so we won't be adding the cookie.
			return
		}
		cookieID, err = addCookieID(w)
		if err != nil {
			// This is a server misconfiguration.
			panic("cannot add cookie ID")
		}
	}

	tmplResp, ok := resp.(*safehttp.TemplateResponse)
	if !ok {
		// If it's not a template response, we cannot inject the token.
		// TODO: should this be an error?
		return
	}

	tok := xsrftoken.Generate(it.SecretAppKey, cookieID.Value(), r.URL().Host())
	if tmplResp.FuncMap == nil {
		tmplResp.FuncMap = map[string]interface{}{}
	}
	tmplResp.FuncMap[htmlinject.XSRFTokensDefaultFuncName] = func() string { return tok }
}

// Match returns false since there are no supported configurations.
func (*Interceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}
