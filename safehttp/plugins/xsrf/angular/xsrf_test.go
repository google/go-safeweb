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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"strings"
	"testing"
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
			rec := safehttptest.NewResponseRecorder()
			test.it.Commit(rec.ResponseWriter, req, nil, nil)

			if got, want := rec.Status(), safehttp.StatusOK; got != want {
				t.Errorf("rec.Status(): got %v, want %v", got, want)
			}

			wantCookie := "Path=/; Max-Age=86400; Secure; SameSite=Strict"
			got := map[string][]string(rec.Header())["Set-Cookie"][0]
			if !strings.Contains(got, test.cookie) {
				t.Errorf("Set-Cookie header: %s not present", test.cookie)
			}
			if !strings.Contains(got, wantCookie) {
				t.Errorf("Set-Cookie header: got %q, want defaults %q", got, wantCookie)
			}

			if got, want := rec.Body(), ""; got != want {
				t.Errorf("rec.Body(): got %q want %q", got, want)
			}

		})
	}
}

func TestAddCookieFail(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	rec := safehttptest.NewResponseRecorder()
	it := &Interceptor{}
	it.Commit(rec.ResponseWriter, req, nil, nil)

	if wantStatus := safehttp.StatusInternalServerError; rec.Status() != wantStatus {
		t.Errorf("rec.Status(): got %v want %v", rec.Status(), wantStatus)
	}

	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header(): mismatch (-want +got):\n%s", diff)
	}

	if wantBody, gotBody := "Internal Server Error\n", rec.Body(); wantBody != gotBody {
		t.Errorf("response body: got %v, want %v", gotBody, wantBody)
	}
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
			rec := safehttptest.NewResponseRecorder()
			i := Default()
			i.Before(rec.ResponseWriter, test.req, nil)

			if got := rec.Status(); got != test.wantStatus {
				t.Errorf("rec.Status(): got %v, want %v", got, test.wantStatus)
			}

			if diff := cmp.Diff(test.wantHeader, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := rec.Body(); got != test.wantBody {
				t.Errorf("rec.Body(): got %q want %q", got, test.wantBody)
			}
		})
	}
}
