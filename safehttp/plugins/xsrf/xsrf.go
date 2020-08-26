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
	"github.com/google/go-safeweb/safehttp"
	"golang.org/x/net/xsrftoken"
)

const (
	// TokenKey is the form key used when sending the token as part of POST
	// request.
	TokenKey = "xsrf-token"
)

// UserIdentifier provides the web application users' identifiers,
// needed in generating the XSRF token.
type UserIdentifier interface {
	// UserID returns the user's identifier based on the
	// safehttp.IncomingRequest received.
	UserID(*safehttp.IncomingRequest) (string, error)
}

// Interceptor implements XSRF protection. It requires an application key and a
// storage service. The appKey uniquely identifies each registered service and
// should have high entropy. The storage service supports retrieving ID's of the
// application's users. Both the appKey and user ID are used in the XSRF
// token generation algorithm.
type Interceptor struct {
	AppKey     string
	Identifier UserIdentifier
}

type tokenCtxKey struct{}

// Token returns the XSRF token from the safehttp.IncomingRequest Context, if
// present, and, otherwise, returns an empty string.
func Token(r *safehttp.IncomingRequest) string {
	tok := r.Context().Value(tokenCtxKey{})
	if tok == nil {
		return ""
	}
	return tok.(string)
}

// Before should be executed before directing the safehttp.IncomingRequest to
// the handler to ensure it is not part of the Cross-Site Request
// Forgery. In case of state changing methods, it checks for the
// presence of an xsrf-token in the request body and validates it based on the
// userID associated with the request. It also adds a cryptographically safe
// XSRF token to the safehttp.IncomingRequest context that will be subsequently
// injected in HTML as hidden input to forms.
func (i *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg interface{}) safehttp.Result {
	userID, err := i.Identifier.UserID(r)
	if err != nil {
		return w.ClientError(safehttp.StatusUnauthorized)
	}

	tok := xsrftoken.Generate(i.AppKey, userID, r.URL.String())
	r.SetContext(context.WithValue(r.Context(), tokenCtxKey{}, tok))

	stateChangingMethods := map[string]bool{
		safehttp.MethodPost:  true,
		safehttp.MethodPatch: true,
	}
	if m := r.Method(); stateChangingMethods[m] {
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
	}

	return safehttp.Result{}
}
