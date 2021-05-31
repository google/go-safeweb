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

// +build go1.16

package auth

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/examples/sample-application/secure/responses"
	"github.com/google/go-safeweb/examples/sample-application/storage"
)

const sessionCookie = "SESSION"

var unauthMsg = template.MustParseAndExecuteToHTML(`Please <a href="/">login</a> before visiting this page.`)

// Interceptor is an auth (access control) interceptor.
//
// It showcases how safehttp.Interceptor could be implement to provide custom
// security features. See https://pkg.go.dev/github.com/google/go-safeweb/safehttp#hdr-Interceptors.
//
// In order to interact with the interceptor, use functions from this package.
// E.g. to clear a user session call ClearSession.
type Interceptor struct {
	DB *storage.DB
}

// Before runs before the request is passed to the handler.
//
// Implementation details: this interceptor uses IncomingRequest's context to
// store user information that's read from a cookie.
func (ip Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	// Identify the user.
	user := ip.userFromCookie(r)
	if user != "" {
		r.SetContext(ctxWithUser(r.Context(), user))
	}

	if _, ok := cfg.(Skip); ok {
		// If the config says we should not perform auth, let's stop executing here.
		return safehttp.NotWritten()
	}

	if user == "" {
		// We have to perform auth, and the user was not identified, bail out.
		return w.WriteError(responses.Error{
			StatusCode: safehttp.StatusUnauthorized,
			Message:    unauthMsg,
		})
	}
	return safehttp.NotWritten()
}

// Commit runs after the handler commited to a response.
//
// Implementation details: the interceptor reads IncomingRequest's context to
// retrieve information about the user and to do what the handler asked it to
// (through ClearSession or CreateSession).
func (ip Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	user := User(r)

	switch ctxSessionAction(r.Context()) {
	case clearSess:
		ip.DB.DelSession(user)
		w.AddCookie(safehttp.NewCookie(sessionCookie, ""))
	case setSess:
		token := ip.DB.GetToken(user)
		w.AddCookie(safehttp.NewCookie(sessionCookie, token))
	default:
		// do nothing
	}
}

func (Interceptor) Match(cfg safehttp.InterceptorConfig) bool {
	_, ok := cfg.(Skip)
	return ok
}

// User retrieves the user.
func User(r *safehttp.IncomingRequest) string {
	return ctxUser(r.Context())
}

func (ip Interceptor) userFromCookie(r *safehttp.IncomingRequest) string {
	sess, err := r.Cookie(sessionCookie)
	if err != nil || sess.Value() == "" {
		return ""
	}
	user, ok := ip.DB.GetUser(sess.Value())
	if !ok {
		return ""
	}
	return user
}

// ClearSession clears the session.
//
// Implementation details: to interact with the interceptor, passes data through
// the IncomingRequest's context.
func ClearSession(r *safehttp.IncomingRequest) {
	r.SetContext(ctxWithSessionAction(r.Context(), clearSess))
}

// CreateSession creates a session.
//
// Implementation details: to interact with the interceptor, passes data through
// the IncomingRequest's context.
func CreateSession(r *safehttp.IncomingRequest, user string) {
	r.SetContext(ctxWithSessionAction(r.Context(), setSess))
	r.SetContext(ctxWithUser(r.Context(), user))
}

// Skip allows to mark an endpoint to skip auth checks.
//
// Its uses would normally be gated by a security review. You can use the
// https://github.com/google/go-safeweb/blob/master/cmd/bancheck tool to enforce
// this.
type Skip struct{}
