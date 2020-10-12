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
	"github.com/google/go-safeweb/safehttp"
	"golang.org/x/net/xsrftoken"
)

const (
	cookieIDKey = "xsrf-cookie"
	TokenKey    = "xsrf-token"
)

type tokenCtxKey struct{}

var statePreservingMethods = map[string]bool{
	safehttp.MethodGet:     true,
	safehttp.MethodHead:    true,
	safehttp.MethodOptions: true,
}

type Interceptor struct {
	secretAppKey string
	g            Generator
	c            Checker
	i            Injector
}

var _ safehttp.Interceptor = &Interceptor{}

func New(key string, g Generator, c Checker, i Injector) Interceptor {
	return Interceptor{
		secretAppKey: key,
		g:            g,
		c:            c,
		i:            i,
	}
}

func Default(key string) Interceptor {
	return Interceptor{
		secretAppKey: key,
		g:            defaultGenerator{secretAppKey: key},
		c:            defaultChecker{secretAppKey: key},
		i:            defaultInjector{},
	}
}

func Angular(cookieName, headerName string) Interceptor {
	return Interceptor{
		g: angularGenerator{tokenCookieName: cookieName},
		c: angularChecker{
			tokenCookieName: cookieName,
			tokenHeaderName: headerName,
		},
		i: angularInjector{},
	}
}

func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if !statePreservingMethods[r.Method()] {
		return it.c.Check(w, r)
	}
	return safehttp.NotWritten()
}

func (it *Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) safehttp.Result {
	it.g.Generate(w, r)
	if w.Written() {
		return safehttp.Result{}
	}
	tok, err := it.i.Token(r)
	if err != nil {
		w.WriteError(safehttp.StatusInternalServerError)
	}
	return it.i.Inject(resp, tok)
}

type Generator interface {
	Generate(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result
}

type Checker interface {
	Check(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result
}

type Injector interface {
	Inject(resp safehttp.Response, token string) safehttp.Result
	Token(r *safehttp.IncomingRequest) (string, error)
}

type defaultGenerator struct {
	secretAppKey string
}

func (g defaultGenerator) Generate(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	c, err := r.Cookie(cookieIDKey)
	if err != nil {
		if !statePreservingMethods[r.Method()] {
			w.WriteError(safehttp.StatusForbidden)
		}
		buf := make([]byte, 20)
		if _, err := rand.Read(buf); err != nil {
			w.WriteError(safehttp.StatusInternalServerError)
		}
		c = safehttp.NewCookie(cookieIDKey, base64.StdEncoding.EncodeToString(buf))
		c.SetSameSite(safehttp.SameSiteStrictMode)
		err = w.SetCookie(c)
		if err != nil {
			w.WriteError(safehttp.StatusInternalServerError)
		}
	}
	tok := xsrftoken.Generate(g.secretAppKey, c.Value(), r.URL.Path())
	r.SetContext(context.WithValue(r.Context(), tokenCtxKey{}, tok))
	return safehttp.NotWritten()
}

type defaultChecker struct {
	secretAppKey string
}

func (c defaultChecker) Check(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	cookie, err := r.Cookie(cookieIDKey)
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

	if !xsrftoken.Valid(tok, c.secretAppKey, cookie.Value(), r.URL.Path()) {
		return w.WriteError(safehttp.StatusForbidden)
	}
	return safehttp.NotWritten()
}

type defaultInjector struct{}

func (i defaultInjector) Token(r *safehttp.IncomingRequest) (string, error) {
	tok := r.Context().Value(tokenCtxKey{})
	if tok == nil {
		return "", errors.New("xsrf token not found")
	}
	return tok.(string), nil
}

func (i defaultInjector) Inject(resp safehttp.Response, token string) safehttp.Result {

	tmplResp, ok := resp.(safehttp.TemplateResponse)
	if !ok {
		return safehttp.NotWritten()
	}
	// TODO(maramihali@): Change the key when function names are exported by
	// htmlinject
	// TODO: what should happen if the XSRFToken key is not present in the
	// tr.FuncMap?
	tmplResp.FuncMap["XSRFToken"] = func() string { return token }
	return safehttp.NotWritten()
}

type angularGenerator struct {
	tokenCookieName string
}

func (g angularGenerator) Generate(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	c, err := r.Cookie(g.tokenCookieName)
	if err == nil {
		return safehttp.NotWritten()
	}
	if r.Method() != safehttp.MethodGet {
		return w.WriteError(safehttp.StatusUnauthorized)
	}
	tok := make([]byte, 20)
	if _, err := rand.Read(tok); err != nil {
		return w.WriteError(safehttp.StatusInternalServerError)
	}
	c = safehttp.NewCookie(g.tokenCookieName, base64.StdEncoding.EncodeToString(tok))

	c.SetSameSite(safehttp.SameSiteStrictMode)
	c.SetPath("/")
	// Set the duration of the token cookie to 24 hours.
	c.SetMaxAge(86400)
	// Needed in order to make the cookie accessible by JavaScript
	// running on the user's domain.
	c.DisableHTTPOnly()

	if err := w.SetCookie(c); err != nil {
		return w.WriteError(safehttp.StatusInternalServerError)
	}
	return safehttp.NotWritten()
}

type angularChecker struct {
	tokenCookieName string
	tokenHeaderName string
}

func (c angularChecker) Check(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	tokCookie, err := r.Cookie(c.tokenCookieName)
	if err != nil {
		return w.WriteError(safehttp.StatusForbidden)
	}
	tok := r.Header.Get(c.tokenHeaderName)
	if tok == "" {
		return w.WriteError(safehttp.StatusUnauthorized)
	}

	if tok != tokCookie.Value() {
		return w.WriteError(safehttp.StatusForbidden)
	}
	return safehttp.NotWritten()
}

type angularInjector struct{}

func (i angularInjector) Token(_ *safehttp.IncomingRequest) (string, error) {
	return "", nil
}

func (i angularInjector) Inject(_ safehttp.Response, _ string) safehttp.Result {
	return safehttp.NotWritten()
}
