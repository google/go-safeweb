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

package xsrfangular

import (
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

const (
	cookieName = "XSRF-TOKEN"
	headerName = "X-XSRF-TOKEN"
)

func TestAddCookie(t *testing.T) {
	tests := []struct {
		name, cookie string
		it           *Interceptor
	}{
		{
			name:   "Default interceptor",
			it:     Default(),
			cookie: cookieName,
		},
		{
			name: "Custom interceptor",
			it: &Interceptor{
				TokenCookieName: "FOO-TOKEN",
				TokenHeaderName: "X-FOO-TOKEN",
			},
			cookie: "FOO-TOKEN",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
			fakeRW, _ := safehttptest.NewFakeResponseWriter()
			test.it.Commit(fakeRW, req, nil, nil)

			if len(fakeRW.Cookies) != 1 {
				t.Errorf("len(Cookies) = %v, want 1", len(fakeRW.Cookies))
			}

			if got, want := fakeRW.Cookies[0].String(), "Path=/; Max-Age=86400; Secure; SameSite=Strict"; !strings.Contains(got, want) {
				t.Errorf("XSRF cookie got %q, want to contain %q", got, want)
			}
		})
	}
}

func TestAddCookieFail(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	fakeRW, _ := safehttptest.NewFakeResponseWriter()
	it := &Interceptor{}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	it.Commit(fakeRW, req, nil, nil)
}

func TestPostProtection(t *testing.T) {
	tests := []struct {
		name       string
		req        *safehttp.IncomingRequest
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name: "Same cookie and header",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", nil)
				req.Header.Set("Cookie", cookieName+"="+"1234")
				req.Header.Set(headerName, "1234")
				return req
			}(),
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name: "Different cookie and header",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", nil)
				req.Header.Set("Cookie", cookieName+"="+"5768")
				req.Header.Set(headerName, "1234")
				return req
			}(),
			wantStatus: safehttp.StatusUnauthorized,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Unauthorized\n",
		},
		{
			name: "Missing header",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", nil)
				req.Header.Set("Cookie", cookieName+"="+"1234")
				return req
			}(),
			wantStatus: safehttp.StatusUnauthorized,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Unauthorized\n",
		},
		{
			name: "Missing cookie",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", nil)
				req.Header.Set(headerName, "1234")
				return req
			}(),
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			i := Default()
			i.Before(fakeRW, test.req, nil)

			if got := rr.Code; got != int(test.wantStatus) {
				t.Errorf("rr.Code: got %v, want %v", got, test.wantStatus)
			}
		})
	}
}
