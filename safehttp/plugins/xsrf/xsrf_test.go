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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"golang.org/x/net/xsrftoken"
	"strings"
	"testing"
)

type userIdentifier struct{}

func (userIdentifier) UserID(r *safehttp.IncomingRequest) (string, error) {
	return "1234", nil
}

var (
	formTokenTests = []struct {
		name, userID, actionID, wantBody string
		wantStatus                       safehttp.StatusCode
		wantHeader                       map[string][]string
	}{
		{
			name:       "Valid token",
			userID:     "1234",
			actionID:   "POST /pizza",
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Invalid actionID in token generation",
			userID:     "1234",
			actionID:   "HEAD /pizza",
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
			actionID:   "POST /pizza",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
	}
)

func TestTokenPost(t *testing.T) {
	for _, test := range formTokenTests {
		t.Run(test.name, func(t *testing.T) {
			rec := safehttptest.NewResponseRecorder()
			tok := xsrftoken.Generate("testAppKey", test.userID, test.actionID)
			req := safehttptest.NewRequest(safehttp.MethodPost, "https://foo.com/pizza", strings.NewReader(TokenKey+"="+tok))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			i := Interceptor{AppKey: "testAppKey", Identifier: userIdentifier{}}
			i.Before(rec.ResponseWriter, req, nil)

			if got := rec.Status(); got != test.wantStatus {
				t.Errorf("response status: got %v, want %v", got, test.wantStatus)
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

func TestTokenMultipart(t *testing.T) {
	for _, test := range formTokenTests {
		t.Run(test.name, func(t *testing.T) {
			rec := safehttptest.NewResponseRecorder()
			tok := xsrftoken.Generate("testAppKey", test.userID, test.actionID)
			b := "--123\r\n" +
				"Content-Disposition: form-data; name=\"xsrf-token\"\r\n" +
				"\r\n" +
				tok + "\r\n" +
				"--123--\r\n"
			req := safehttptest.NewRequest(safehttp.MethodPost, "https://foo.com/pizza", strings.NewReader(b))
			req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

			i := Interceptor{AppKey: "testAppKey", Identifier: userIdentifier{}}
			i.Before(rec.ResponseWriter, req, nil)

			if got := rec.Status(); got != test.wantStatus {
				t.Errorf("response status: got %v, want %v", got, test.wantStatus)
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
		rec := safehttptest.NewResponseRecorder()

		i := Interceptor{AppKey: "testAppKey", Identifier: userIdentifier{}}
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

func TestBeforeTokenInRequestContext(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{AppKey: "testAppKey", Identifier: userIdentifier{}}
	i.Before(rec.ResponseWriter, req, nil)

	tok, err := Token(req)
	if tok == "" {
		t.Error(`Token(req): got "", want token`)
	}
	if err != nil {
		t.Errorf("Token(req): got %v, want nil", err)
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

}

func TestTokenInRequestContext(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	req.SetContext(context.WithValue(req.Context(), tokenCtxKey{}, "pizza"))

	got, err := Token(req)
	if want := "pizza"; want != got {
		t.Errorf("Token(req): got %v, want %v", got, want)
	}
	if err != nil {
		t.Errorf("Token(req): got %v, want nil", err)
	}
}

func TestMissingTokenInRequestContext(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	req.SetContext(context.Background())

	got, err := Token(req)
	if want := ""; want != got {
		t.Errorf("Token(req): got %v, want %v", got, want)
	}
	if err == nil {
		t.Error("Token(req): got nil, want error")
	}
}
