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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/go-safeweb/safehttp"
	"golang.org/x/net/xsrftoken"
)

var statePreservingMethods = map[string]bool{
	safehttp.MethodGet:     true,
	safehttp.MethodHead:    true,
	safehttp.MethodOptions: true,
}

type Interceptor struct {
	secretAppKey string
	c            Checker
	i            Injector
}

var _ safehttp.Interceptor = &Interceptor{}

func New(key string, c Checker, i Injector) Interceptor {
	return Interceptor{
		secretAppKey: key,
		c:            c,
		i:            i,
	}
}

func Default(key string) Interceptor {
	return Interceptor{
		secretAppKey: key,
		c: defaultChecker{
			secretAppKey: key,
			cookieIDKey:  "xsrf-cookie",
			tokenKey:     "xsrf-token",
		},
		i: defaultInjector{
			secretAppKey: key,
			cookieIDKey:  "xsrf-cookie",
		},
	}
}

func Angular(cookieName, headerName string) Interceptor {
	return Interceptor{
		c: angularChecker{
			tokenCookieName: cookieName,
			tokenHeaderName: headerName,
		},
		i: angularInjector{
			tokenCookieName: cookieName,
		},
	}
}

func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if statePreservingMethods[r.Method()] {
		return safehttp.NotWritten()
	}
	tok, code := it.c.Retrieve(r)
	if code != safehttp.StatusOK {
		return w.WriteError(code)
	}
	code = it.c.Validate(r, tok)
	if code != safehttp.StatusOK {
		return w.WriteError(code)
	}
	return safehttp.NotWritten()
}

func (it *Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {
	err := it.i.Inject(resp, w, r)
	if err != nil {
		return w.WriteError(safehttp.StatusInternalServerError)
	}
	return safehttp.Result{}
}

type Checker interface {
	Retrieve(r *safehttp.IncomingRequest) (string, safehttp.StatusCode)
	Validate(r *safehttp.IncomingRequest, token string) safehttp.StatusCode
}

type Injector interface {
	Inject(resp safehttp.Response, w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) error
}
s
type defaultChecker struct {
	secretAppKey string
	cookieIDKey  string
	tokenKey     string
}

func (c defaultChecker) Retrieve(r *safehttp.IncomingRequest) (string, safehttp.StatusCode) {
	f, err := r.PostForm()
	if err != nil {
		// We fallback to checking whether the form is multipart. Both types
		// are valid in an incoming request as long as the XSRF token is
		// present.
		mf, err := r.MultipartForm(32 << 20)
		if err != nil {
			return "", safehttp.StatusBadRequest
		}
		f = &mf.Form
	}

	tok := f.String(c.tokenKey, "")
	if f.Err() != nil || tok == "" {
		return "", safehttp.StatusUnauthorized
	}

	return tok, safehttp.StatusOK
}

func (c defaultChecker) Validate(r *safehttp.IncomingRequest, token string) safehttp.StatusCode {
	cookie, err := r.Cookie(c.cookieIDKey)
	if err != nil {
		return safehttp.StatusForbidden
	}

	if !xsrftoken.Valid(token, c.secretAppKey, cookie.Value(), r.URL.Path()) {
		return safehttp.StatusForbidden
	}
	return safehttp.StatusOK
}

type defaultInjector struct {
	cookieIDKey  string
	secretAppKey string
}

func (i defaultInjector) Inject(resp safehttp.Response, w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) error {
	c, err := r.Cookie(i.cookieIDKey)
	if err != nil {
		buf := make([]byte, 20)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("crypto/rand.Read: %v", err)
		}
		c = safehttp.NewCookie(i.cookieIDKey, base64.StdEncoding.EncodeToString(buf))
		c.SetSameSite(safehttp.SameSiteStrictMode)

		if err := w.SetCookie(c); err != nil {
			return err
		}
	}

	tok := xsrftoken.Generate(i.secretAppKey, c.Value(), r.URL.Path())

	tmplResp, ok := resp.(safehttp.TemplateResponse)
	if !ok {
		return nil
	}

	// TODO(maramihali@): Change the key when function names are exported by
	// htmlinject
	// TODO: what should happen if the XSRFToken key is not present in the
	// tr.FuncMap?
	tmplResp.FuncMap["XSRFToken"] = func() string { return tok }
	return nil
}

type angularChecker struct {
	tokenCookieName string
	tokenHeaderName string
}

func (c angularChecker) Retrieve(r *safehttp.IncomingRequest) (string, safehttp.StatusCode) {
	tok := r.Header.Get(c.tokenHeaderName)
	if tok == "" {
		return "", safehttp.StatusUnauthorized
	}
	return tok, safehttp.StatusOK
}

func (c angularChecker) Validate(r *safehttp.IncomingRequest, tok string) safehttp.StatusCode {
	cookie, err := r.Cookie(c.tokenCookieName)
	if err != nil {
		return safehttp.StatusForbidden
	}

	if tok != cookie.Value() {
		return safehttp.StatusUnauthorized
	}
	return safehttp.StatusOK
}

type angularInjector struct {
	tokenCookieName string
}

func (i angularInjector) Inject(resp safehttp.Response, w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) error {
	c, err := r.Cookie(i.tokenCookieName)
	if err == nil {
		return nil
	}
	tok := make([]byte, 20)
	if _, err := rand.Read(tok); err != nil {
		return fmt.Errorf("crypto/rand.Read: %v", err)
	}
	c = safehttp.NewCookie(i.tokenCookieName, base64.StdEncoding.EncodeToString(tok))

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
