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

package xsrf

import (
	"context"
	"errors"
	"github.com/google/go-safeweb/safehttp"
	"golang.org/x/net/xsrftoken"
)

const (
	// TokenKey is the form key used when sending the token as part of POST
	// request.
	TokenKey = "xsrf-token"
)

var statePreservingMethods = map[string]bool{
	safehttp.MethodGet:     true,
	safehttp.MethodHead:    true,
	safehttp.MethodOptions: true,
}

// UserIdentifier provides the web application users' identifiers,
// needed in generating the XSRF token.
type UserIdentifier interface {
	// UserID returns the user's identifier based on the
	// safehttp.IncomingRequest received.
	UserID(*safehttp.IncomingRequest) (string, error)
}

// Interceptor implements XSRF protection.
type Interceptor struct {
	// AppKey uniquely identifies each registered service and should have high
	// entropy as it is use in generating the XSRF token.
	AppKey string
	// Identifier supports retrieving the user ID of application's user based on
	// incoming requests. This is needed in generating the XSRF token.
	Identifier UserIdentifier
}

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

// Before should be executed before directing the safehttp.IncomingRequest to
// the handler to ensure it is not part of the Cross-Site Request
// Forgery attack.
//
// In case of state changing requests (all except GET, HEAD and OPTIONS), it
// checks for the presence of an XSRF token in the request and validates it
// based on the user ID associated with the request.
//
// For authorized requests, it adds a cryptographically safe XSRF token to the
// incoming request. It can be later extracted using Token.
func (i *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg interface{}) safehttp.Result {
	userID, err := i.Identifier.UserID(r)
	if err != nil {
		return w.ClientError(safehttp.StatusUnauthorized)
	}

	if m := r.Method(); statePreservingMethods[m] {
		tok := xsrftoken.Generate(i.AppKey, userID, r.URL.String())
		r.SetContext(context.WithValue(r.Context(), tokenCtxKey{}, tok))
		return safehttp.Result{}
	}

	f, err := r.PostForm()
	if err != nil {
		mf, err := r.MultipartForm(32 << 20)
		if err != nil {
			return w.ClientError(safehttp.StatusBadRequest)
		}
		f = &mf.Form
	}

	tok := f.String(TokenKey, "")
	if f.Err() != nil || tok == "" {
		return w.ClientError(safehttp.StatusUnauthorized)
	}

	if ok := xsrftoken.Valid(tok, i.AppKey, userID, r.URL.String()); !ok {
		return w.ClientError(safehttp.StatusForbidden)
	}

	tok = xsrftoken.Generate(i.AppKey, userID, r.URL.String())
	r.SetContext(context.WithValue(r.Context(), tokenCtxKey{}, tok))

	return safehttp.Result{}
}
