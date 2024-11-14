// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"testing"

	"github.com/google/go-safeweb/examples/sample-application/storage"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

const TEST_USER = "test"

func TestInterceptorBefore(t *testing.T) {
	tests := []struct {
		name    string
		cfg     safehttp.InterceptorConfig
		hasAuth bool
		want    safehttp.StatusCode
	}{
		{
			name:    "base case, no error",
			hasAuth: true,
			cfg:     nil,
			want:    safehttp.StatusOK,
		},
		{
			name:    "force skip using config",
			hasAuth: true,
			cfg:     Skip{},
			want:    safehttp.StatusOK,
		},
		{
			name:    "missing auth, error",
			hasAuth: false,
			cfg:     nil,
			want:    safehttp.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withUserDB, token := addTestUser(storage.NewDB())
			ip := newTestInterceptor(withUserDB)

			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
			rw, r := safehttptest.NewFakeResponseWriter()

			if tt.hasAuth {
				addTestUserCookie(req, token)
			}

			// Note: "Before" return value is not significant
			ip.Before(rw, req, tt.cfg)

			if got := r.Code; got != int(tt.want) {
				t.Errorf("status code got: %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInterceptorCommit(t *testing.T) {
	tests := []struct {
		name      string
		action    sessionAction
		hasCookie bool
	}{
		{
			name:      "clear session, no error",
			action:    clearSess,
			hasCookie: false,
		}, {
			name:      "set session, no error",
			action:    setSess,
			hasCookie: true,
		},
		{
			name:      "unexpected action, skip",
			action:    sessionAction("unexpected"),
			hasCookie: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withUserDB, _ := addTestUser(storage.NewDB())
			ip := newTestInterceptor(withUserDB)

			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
			rw, r := safehttptest.NewFakeResponseWriter()

			safehttp.FlightValues(req.Context()).Put(userKey, "user")
			safehttp.FlightValues(req.Context()).Put(changeSessKey, tt.action)

			ip.Commit(rw, req, r.Result, nil)

			var token string
			for _, c := range rw.Cookies {
				if c.Name() == sessionCookie {
					token = c.Value()
				}
			}

			if tt.hasCookie == (token == "") {
				t.Errorf("token = %q, want %v", token, tt.hasCookie)
			}
		})
	}
}

func TestInterceptorMatch(t *testing.T) {
	tests := []struct {
		name string
		cfg  safehttp.InterceptorConfig
		want bool
	}{
		{
			name: "basic case, no error",
			cfg:  Skip{},
			want: true,
		},
		{
			name: "no Skip{}, error",
			cfg:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := newTestInterceptor(nil)
			if got := ip.Match(tt.cfg); got != tt.want {
				t.Errorf("Interceptor.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterceptorUserFromCookie(t *testing.T) {
	withUserDB, validToken := addTestUser(storage.NewDB())
	ip := newTestInterceptor(withUserDB)

	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "basic case, no error",
			token: validToken,
			want:  TEST_USER,
		},
		{
			name:  "empty cookie, error",
			token: "",
			want:  "",
		},
		{
			name:  "invalid token, error",
			token: "not_a_valid_token",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
			addTestUserCookie(req, tt.token)

			if got := ip.userFromCookie(req); got != tt.want {
				t.Errorf("Interceptor.userFromCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionManagement(t *testing.T) {
	want := "wanted"
	r := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)

	CreateSession(r, want)
	if got := User(r); got != want {
		t.Errorf("user id got: %q, want %q", got, want)
	}

	ClearSession(r)
	// Note: `ctxSessionAction` already tested inside ctx_test.go
	if got := ctxSessionAction(r.Context()); got != clearSess {
		t.Errorf("no clearSess action found in context after ClearSession")
	}
}

func addTestUserCookie(r *safehttp.IncomingRequest, v string) {
	r.Header.Add("Cookie", safehttp.NewCookie(sessionCookie, v).String())
}

func newTestInterceptor(db *storage.DB) Interceptor {
	if db == nil {
		db = storage.NewDB()
	}
	return Interceptor{
		DB: db,
	}
}

func addTestUser(db *storage.DB) (*storage.DB, string) {
	token := (*db).GetToken(TEST_USER)
	return db, token
}
