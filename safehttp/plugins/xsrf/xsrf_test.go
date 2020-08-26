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
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"golang.org/x/net/xsrftoken"
	"strings"
	"testing"
)

type userIdentifier struct{}

func (userIdentifier) UserID(r *safehttp.IncomingRequest) (string, error) {
	return "1234", nil
}

func TestTokenPost(t *testing.T) {
	methods := []string{safehttp.MethodPost, safehttp.MethodPatch}
	tests := []struct {
		name       string
		userID     string
		actionID   string
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid token",
			userID:     "1234",
			actionID:   "https://foo.com/pizza",
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Token mismatch for invalid host",
			userID:     "1234",
			actionID:   "https://bar.com/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Token mismatch for invalid path",
			userID:     "1234",
			actionID:   "https://foo.com/spaghetti",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Token mismatch for invalid protocol",
			userID:     "1234",
			actionID:   "http://foo.com/spaghetti",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Token mismatch for invalid userID",
			userID:     "5678",
			actionID:   "http://foo.com/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}
	for _, method := range methods {
		for _, test := range tests {
			t.Run(method+test.name, func(t *testing.T) {
				i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
				tok := xsrftoken.Generate("xsrf", test.userID, test.actionID)
				req := safehttptest.NewRequest(method, "https://foo.com/pizza", strings.NewReader(xsrf.TokenKey+"="+tok))
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
			})
		}
	}
}

func TestTokenMultipart(t *testing.T) {
	methods := []string{safehttp.MethodPost, safehttp.MethodPatch}
	tests := []struct {
		name       string
		userID     string
		actionID   string
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid token",
			userID:     "1234",
			actionID:   "https://foo.com/pizza",
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Token mismatch for invalid host",
			userID:     "1234",
			actionID:   "https://bar.com/pizza",
			wantStatus: 403,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Token mismatch for invalid path",
			userID:     "1234",
			actionID:   "https://foo.com/spaghetti",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Token mismatch for invalid protocol",
			userID:     "1234",
			actionID:   "http://foo.com/spaghetti",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Token mismatch for invalid user ID",
			userID:     "5678",
			actionID:   "http://foo.com/pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}
	for _, method := range methods {
		for _, test := range tests {
			t.Run(method+test.name, func(t *testing.T) {
				i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
				tok := xsrftoken.Generate("xsrf", test.userID, test.actionID)
				b := "--123\r\n" +
					"Content-Disposition: form-data; name=\"xsrf-token\"\r\n" +
					"\r\n" +
					tok + "\r\n" +
					"--123--\r\n"
				req := safehttptest.NewRequest(method, "https://foo.com/pizza", strings.NewReader(b))
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
			})
		}
	}
}

func TestMissingTokenInBody(t *testing.T) {
	tests := []struct {
		name string
		req  *safehttp.IncomingRequest
	}{
		{
			name: "Missing token in POST request with form",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader("foo=bar"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
		},
		{
			name: "Missing token in PATCH request with form",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPatch, "/", strings.NewReader("foo=bar"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
		},
		{
			name: "Missing token in POST request with multipart form",
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
		},
		{
			name: "Missing token in PATCH request with multipart form",
			req: func() *safehttp.IncomingRequest {
				b := "--123\r\n" +
					"Content-Disposition: form-data; name=\"foo\"\r\n" +
					"\r\n" +
					"bar\r\n" +
					"--123--\r\n"
				req := safehttptest.NewRequest(safehttp.MethodPatch, "/", strings.NewReader(b))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
		},
	}
	for _, test := range tests {
		i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
		rec := safehttptest.NewResponseRecorder()

		i.Before(rec.ResponseWriter, test.req, nil)

		if want, got := safehttp.StatusUnauthorized, rec.Status(); got != want {
			t.Errorf("response status: got %v, want %v", got, want)
		}
		wantHeaders := map[string][]string{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		}
		if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
			t.Errorf("rw.header mismatch (-want +got):\n%s", diff)
		}
		if want, got := "Unauthorized\n", rec.Body(); got != want {
			t.Errorf("response body: got %q want %q", got, want)
		}
	}
}

func TestTokenRequestContext(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		actionID string
		match    bool
	}{
		{
			name:     "Valid token",
			userID:   "1234",
			actionID: "https://foo.com/pizza",
			match:    true,
		},
		{
			name:     "Token mismatch for invalid host",
			userID:   "1234",
			actionID: "https://bar.com/pizza",
			match:    false,
		},
		{
			name:     "Token mismatch for invalid path",
			userID:   "1234",
			actionID: "https://foo.com/spaghetti",
			match:    false,
		},
		{
			name:     "Token mismatch for invalid protocol",
			userID:   "1234",
			actionID: "http://foo.com/spaghetti",
			match:    false,
		},
		{
			name:     "Token mismatch for invalid user ID",
			userID:   "5678",
			actionID: "http://foo.com/pizza",
			match:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i := xsrf.Interceptor{AppKey: "xsrf", Identifier: userIdentifier{}}
			req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
			rec := safehttptest.NewResponseRecorder()

			i.Before(rec.ResponseWriter, req, nil)

			gotTok := xsrf.Token(req)
			if ok := xsrftoken.Valid(gotTok, "xsrf", test.userID, test.actionID); ok != test.match {
				t.Errorf("xsrf.Token(req): invalid token")
			}

			if want, got := safehttp.StatusOK, safehttp.StatusCode(rec.Status()); want != got {
				t.Errorf("response status: got %v, want %v", got, want)
			}
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
				t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
			}
			if want, got := "", rec.Body(); got != want {
				t.Errorf("response body: got %q want %q", got, want)
			}
		})
	}

}

func TestMissingTokenRequestContext(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	req.SetContext(context.Background())

	if want, got := "", xsrf.Token(req); want != got {
		t.Errorf("xsrf.Token(req): got %v, want %v", got, want)
	}
}
