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

var (
	formTokenTests = []struct {
		name, cookieVal, actionID, wantBody string
		wantStatus                          safehttp.StatusCode
		wantHeader                          map[string][]string
	}{
		{
			name:       "Valid token",
			cookieVal:  "abcdef",
			actionID:   "/pizza",
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{},
			wantBody:   "",
		},
		{
			name:       "Invalid actionID in token generation",
			cookieVal:  "abcdef",
			actionID:   "/spaghetti",
			wantStatus: safehttp.StatusForbidden,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Forbidden\n",
		},
		{
			name:       "Invalid cookie value in token generation",
			cookieVal:  "evilvalue",
			actionID:   "/pizza",
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
			tok := xsrftoken.Generate("testSecretAppKey", test.cookieVal, test.actionID)
			req := safehttptest.NewRequest(safehttp.MethodPost, "https://foo.com/pizza", strings.NewReader(TokenKey+"="+tok))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Cookie", cookieIDKey+"=abcdef")

			i := Interceptor{SecretAppKey: "testSecretAppKey"}
			i.Before(rec.ResponseWriter, req, nil)

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

func TestTokenMultipart(t *testing.T) {
	for _, test := range formTokenTests {
		t.Run(test.name, func(t *testing.T) {
			rec := safehttptest.NewResponseRecorder()
			tok := xsrftoken.Generate("testSecretAppKey", test.cookieVal, test.actionID)
			b := "--123\r\n" +
				"Content-Disposition: form-data; name=\"xsrf-token\"\r\n" +
				"\r\n" +
				tok + "\r\n" +
				"--123--\r\n"
			req := safehttptest.NewRequest(safehttp.MethodPost, "https://foo.com/pizza", strings.NewReader(b))
			req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
			req.Header.Set("Cookie", cookieIDKey+"=abcdef")

			i := Interceptor{SecretAppKey: "testSecretAppKey"}
			i.Before(rec.ResponseWriter, req, nil)

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

func TestMalformedForm(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodPost, "https://foo.com/pizza", nil)
	req.Header.Set("Content-Type", "wrong")
	req.Header.Set("Cookie", cookieIDKey+"=abcdef")

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Before(rec.ResponseWriter, req, nil)

	if want, got := safehttp.StatusBadRequest, rec.Status(); got != want {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}
	if want, got := "Bad Request\n", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
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
				req.Header.Set("Cookie", cookieIDKey+"=abcdef")
				return req
			}(),
		},
		{
			name: "Missing token in PATCH request with form",
			req: func() *safehttp.IncomingRequest {
				req := safehttptest.NewRequest(safehttp.MethodPatch, "/", strings.NewReader("foo=bar"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Cookie", cookieIDKey+"=abcdef")
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
				req.Header.Set("Cookie", cookieIDKey+"=abcdef")
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
				req.Header.Set("Cookie", cookieIDKey+"=abcdef")
				return req
			}(),
		},
	}
	for _, test := range tests {
		rec := safehttptest.NewResponseRecorder()

		i := Interceptor{SecretAppKey: "testSecretAppKey"}
		i.Before(rec.ResponseWriter, test.req, nil)

		if want, got := safehttp.StatusUnauthorized, rec.Status(); got != want {
			t.Errorf("rec.Status(): got %v, want %v", got, want)
		}
		wantHeaders := map[string][]string{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		}
		if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
			t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
		}
		if want, got := "Unauthorized\n", rec.Body(); got != want {
			t.Errorf("rec.Body(): got %q want %q", got, want)
		}
	}
}

func TestBeforeTokenInRequestContext(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	req.Header.Set("Cookie", cookieIDKey+"=abcdef")

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Before(rec.ResponseWriter, req, nil)

	tok, err := Token(req)
	if tok == "" {
		t.Error(`Token(req): got "", want token`)
	}
	if err != nil {
		t.Errorf("Token(req): got %v, want nil", err)
	}

	if want, got := safehttp.StatusOK, rec.Status(); want != got {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}
	if want, got := "", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
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

func TestMissingCookieInGetRequest(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Before(rec.ResponseWriter, req, nil)

	if want, got := safehttp.StatusOK, rec.Status(); want != got {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	tokCookieDefaults := "HttpOnly; Secure; SameSite=Strict"
	got := map[string][]string(rec.Header())["Set-Cookie"][0]
	if got == "" {
		t.Error("rec.Header(): expected Set-Cookie header to be set")
	}
	if !strings.Contains(got, tokCookieDefaults) {
		t.Errorf("Set-Cookie header: got %s, want defaults %s", got, tokCookieDefaults)
	}
	if want, got := "", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
	}
}

func TestMissingCookiePostRequest(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader("foo=bar"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Before(rec.ResponseWriter, req, nil)

	if want, got := safehttp.StatusForbidden, rec.Status(); got != want {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}
	if want, got := "Forbidden\n", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
	}
}

func TestCommitToken(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	req.SetContext(context.WithValue(req.Context(), tokenCtxKey{}, "pizza"))

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	tr := safehttp.TemplateResponse{FuncMap: map[string]interface{}{}}
	i.Commit(rec.ResponseWriter, req, tr, nil)

	tok, ok := tr.FuncMap["XSRFToken"]
	if !ok {
		t.Fatal(`tr.FuncMap["XSRFToken"] not found`)
	}

	fn, ok := tok.(func() string)
	if !ok {
		t.Errorf(`tr.FuncMap["XSRFToken"]: got %T, want "func() string"`, fn)
	}
	if want, got := "pizza", fn(); want != got {
		t.Errorf(`tr.FuncMap["XSRFToken"](): got %q, want %q`, got, want)
	}

	if want, got := safehttp.StatusOK, rec.Status(); want != got {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}

	if want, got := "", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
	}
}

func TestCommitMissingToken(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)
	req.SetContext(context.Background())

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	tr := safehttp.TemplateResponse{FuncMap: map[string]interface{}{}}
	i.Commit(rec.ResponseWriter, req, tr, nil)

	wantFuncMap := map[string]interface{}{}
	if diff := cmp.Diff(wantFuncMap, tr.FuncMap); diff != "" {
		t.Errorf("tr.FuncMap: mismatch (-want +got):\n%s", diff)
	}

	if want, got := safehttp.StatusInternalServerError, rec.Status(); got != want {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	wantHeaders := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}

	if diff := cmp.Diff(wantHeaders, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}

	if want, got := "Internal Server Error\n", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
	}
}

func TestCommitNotTemplateResponse(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Commit(rec.ResponseWriter, req, safehttp.NoContentResponse{}, nil)

	if want, got := safehttp.StatusOK, rec.Status(); want != got {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}

	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rec.Header())); diff != "" {
		t.Errorf("rec.Header() mismatch (-want +got):\n%s", diff)
	}

	if want, got := "", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
	}
}

func TestBeforeCommit(t *testing.T) {
	rec := safehttptest.NewResponseRecorder()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	tr := safehttp.TemplateResponse{FuncMap: map[string]interface{}{}}
	i.Before(rec.ResponseWriter, req, nil)
	i.Commit(rec.ResponseWriter, req, tr, nil)

	tok, ok := tr.FuncMap["XSRFToken"]
	if !ok {
		t.Fatal(`tr.FuncMap["XSRFToken"] not found`)
	}

	fn, ok := tok.(func() string)
	if !ok {
		t.Errorf(`tr.FuncMap["XSRFToken"]: got %T, want "func() string"`, fn)
	}
	if got := fn(); got == "" {
		t.Errorf(`tr.FuncMap["XSRFToken"](): got %q, want token`, got)
	}

	if want, got := safehttp.StatusOK, rec.Status(); want != got {
		t.Errorf("rec.Status(): got %v, want %v", got, want)
	}
	if want, got := "", rec.Body(); got != want {
		t.Errorf("rec.Body(): got %q want %q", got, want)
	}

}
