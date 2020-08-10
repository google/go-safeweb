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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"strings"
	"testing"
)

type testUserIDStorage struct{}

func (testUserIDStorage) GetUserID() (string, error) {
	return "potato", nil
}

func TestXSRFTokenPost(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		host       string
		path       string
		wantStatus int
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid token",
			target:     "http://foo.com/pizza",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: 200,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Invalid host in token generation",
			target:     "http://foo.com/pizza",
			host:       "bar.com",
			path:       "/pizza",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid path in token generation",
			target:     "http://foo.com/pizza",
			host:       "foo.com",
			path:       "/spaghetti",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		//TODO(@mihalimara22): Add tests for invalid user ID once
		//UserIDStorage.GetUserID receives a parameter
	}
	for _, test := range tests {
		p := xsrf.NewPlugin("1234", testUserIDStorage{})
		tok, err := p.GenerateToken(test.host, test.path)
		if err != nil {
			t.Fatalf("p.GenerateToken: got %v, want nil", err)
		}
		req := safehttptest.NewRequest("POST", test.target, strings.NewReader(xsrf.TokenKey+"="+tok))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := safehttptest.NewResponseRecorder()
		p.Before(rec.ResponseWriter, req)
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
		target     string
		host       string
		path       string
		wantStatus int
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid token",
			target:     "http://foo.com/pizza",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: 200,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Invalid host in token generation",
			target:     "http://foo.com/pizza",
			host:       "bar.com",
			path:       "/pizza",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid path in token generation",
			target:     "http://foo.com/pizza",
			host:       "foo.com",
			path:       "/spaghetti",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		//TODO(@mihalimara22): Add tests for invalid user ID once
		//UserIDStorage.GetUserID receives a parameter
	}
	for _, test := range tests {
		p := xsrf.NewPlugin("1234", testUserIDStorage{})
		tok, err := p.GenerateToken(test.host, test.path)
		if err != nil {
			t.Fatalf("p.GenerateToken: got %v, want nil", err)
		}
		reqBody := "--123\r\n" +
			"Content-Disposition: form-data; name=\"xsrf-token\"\r\n" +
			"\r\n" +
			tok + "\r\n" +
			"--123--\r\n"
		req := safehttptest.NewRequest("POST", test.target, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
		rec := safehttptest.NewResponseRecorder()
		p.Before(rec.ResponseWriter, req)
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
		wantStatus int
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name: "Missing token in POST request",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest("POST", "/", strings.NewReader("foo=bar"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			wantStatus: 401,
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
				req := safehttptest.NewRequest("POST", "/", strings.NewReader(b))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			wantStatus: 401,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Unauthorized\n",
		},
	}
	for _, test := range tests {
		p := xsrf.NewPlugin("1234", testUserIDStorage{})
		rec := safehttptest.NewResponseRecorder()
		p.Before(rec.ResponseWriter, test.req)
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
