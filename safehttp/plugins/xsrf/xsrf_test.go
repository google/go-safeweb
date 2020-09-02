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

package xsrf_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"golang.org/x/net/xsrftoken"
)

type userIdentifier struct{}

func (userIdentifier) UserID(r *safehttp.IncomingRequest) (string, error) {
	return "1234", nil
}

func TestXSRFTokenPost(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		path       string
		userID     string
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid token",
			userID:     "1234",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Invalid host in token generation",
			userID:     "1234",
			host:       "bar.com",
			path:       "/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid path in token generation",
			userID:     "1234",
			host:       "foo.com",
			path:       "spaghetti",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid userID in token generation",
			userID:     "5678",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}
	for _, test := range tests {
		i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
		tok := xsrftoken.Generate("xsrf", test.userID, test.host+test.path)
		req := safehttptest.NewRequest(safehttp.MethodPost, "http://foo.com/pizza", strings.NewReader(xsrf.TokenKey+"="+tok))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rec := safehttptest.NewResponseRecorder()
		i.Before(rec.ResponseWriter, req, nil)

		if rec.Status() != test.wantStatus {
			t.Errorf("response status: got %v, want %v", rec.Status(), test.wantStatus)
		}
		if diff := cmp.Diff(test.wantHeader, map[string][]string(rec.Header())); diff != "" {
			t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
		}
		if got := rec.Body(); got != test.wantBody {
			t.Errorf("response body: got %q want %q", got, test.wantBody)
		}
	}
}

func TestXSRFTokenMultipart(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		path       string
		userID     string
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid token",
			userID:     "1234",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Invalid host in token generation",
			userID:     "1234",
			host:       "bar.com",
			path:       "/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid path in token generation",
			userID:     "1234",
			host:       "foo.com",
			path:       "spaghetti",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid userID in token generation",
			userID:     "5678",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}
	for _, test := range tests {
		i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
		tok := xsrftoken.Generate("xsrf", test.userID, test.host+test.path)
		b := "--123\r\n" +
			"Content-Disposition: form-data; name=\"xsrf-token\"\r\n" +
			"\r\n" +
			tok + "\r\n" +
			"--123--\r\n"
		req := safehttptest.NewRequest(safehttp.MethodPost, "http://foo.com/pizza", strings.NewReader(b))
		req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

		rec := safehttptest.NewResponseRecorder()
		i.Before(rec.ResponseWriter, req, nil)

		if rec.Status() != test.wantStatus {
			t.Errorf("response status: got %v, want %v", rec.Status(), test.wantStatus)
		}
		if diff := cmp.Diff(test.wantHeader, map[string][]string(rec.Header())); diff != "" {
			t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
		}
		if got := rec.Body(); got != test.wantBody {
			t.Errorf("response body: got %q want %q", got, test.wantBody)
		}
	}
}

func TestXSRFMissingToken(t *testing.T) {
	tests := []struct {
		name       string
		req        *safehttp.IncomingRequest
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name: "Missing token in POST request",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader("foo=bar"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
			name: "Missing token in multipart request",
			req: func() *safehttp.IncomingRequest {
				b := "--123\r\n" +
					"Content-Disposition: form-data; name=\"foo\"\r\n" +
					"\r\n" +
					"bar\r\n" +
					"--123--\r\n"
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader(b))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			wantStatus: safehttp.StatusUnauthorized,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Unauthorized\n",
		},
	}
	for _, test := range tests {
		i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
		rec := safehttptest.NewResponseRecorder()
		i.Before(rec.ResponseWriter, test.req, nil)

		if rec.Status() != test.wantStatus {
			t.Errorf("response status: got %v, want %v", rec.Status(), test.wantStatus)
		}
		if diff := cmp.Diff(test.wantHeader, map[string][]string(rec.Header())); diff != "" {
			t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
		}
		if got := rec.Body(); got != test.wantBody {
			t.Errorf("response body: got %q want %q", got, test.wantBody)
		}
	}
}
