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

package xsrfhtml

import (
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"golang.org/x/net/xsrftoken"
)

var (
	formTokenTests = []struct {
		name, cookieVal, host string
		wantStatus            safehttp.StatusCode
	}{
		{
			name:       "Valid token",
			cookieVal:  "abcdef",
			host:       "go.dev",
			wantStatus: safehttp.StatusOK,
		},
		{
			name:       "Invalid host in token generation",
			cookieVal:  "abcdef",
			host:       "google.com",
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name:       "Invalid cookie value in token generation",
			cookieVal:  "evilvalue",
			host:       "go.dev",
			wantStatus: safehttp.StatusForbidden,
		},
	}
)

func TestTokenPost(t *testing.T) {
	for _, test := range formTokenTests {
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			tok := xsrftoken.Generate("testSecretAppKey", test.cookieVal, test.host)
			req := safehttptest.NewRequest(safehttp.MethodPost, "https://go.dev/", strings.NewReader(TokenKey+"="+tok))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Cookie", cookieIDKey+"=abcdef")

			i := Interceptor{SecretAppKey: "testSecretAppKey"}
			i.Before(fakeRW, req, nil)

			if got := rr.Code; got != int(test.wantStatus) {
				t.Errorf("rr.Code: got %v, want %v", got, test.wantStatus)
			}
		})
	}
}

func TestTokenMultipart(t *testing.T) {
	for _, test := range formTokenTests {
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			tok := xsrftoken.Generate("testSecretAppKey", test.cookieVal, test.host)
			b := "--123\r\n" +
				"Content-Disposition: form-data; name=\"xsrf-token\"\r\n" +
				"\r\n" +
				tok + "\r\n" +
				"--123--\r\n"
			req := safehttptest.NewRequest(safehttp.MethodPost, "https://go.dev/", strings.NewReader(b))
			req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
			req.Header.Set("Cookie", cookieIDKey+"=abcdef")

			i := Interceptor{SecretAppKey: "testSecretAppKey"}
			i.Before(fakeRW, req, nil)

			if got := rr.Code; got != int(test.wantStatus) {
				t.Errorf("rr.Code: got %v, want %v", got, test.wantStatus)
			}
		})
	}
}

func TestMalformedForm(t *testing.T) {
	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodPost, "https://foo.com/pizza", nil)
	req.Header.Set("Content-Type", "wrong")
	req.Header.Set("Cookie", cookieIDKey+"=abcdef")

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Before(fakeRW, req, nil)

	if want, got := int(safehttp.StatusBadRequest), rr.Code; got != want {
		t.Errorf("rr.Code: got %v, want %v", got, want)
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
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			i := Interceptor{SecretAppKey: "testSecretAppKey"}
			i.Before(fakeRW, test.req, nil)

			if want, got := safehttp.StatusUnauthorized, rr.Code; got != int(want) {
				t.Errorf("rr.Code: got %v, want %v", got, want)
			}
		})
	}
}

func TestMissingCookieInGetRequest(t *testing.T) {
	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Commit(fakeRW, req, nil, nil)

	if want, got := safehttp.StatusOK, rr.Code; got != int(want) {
		t.Errorf("rr.Code: got %v, want %v", got, want)
	}

	if len(fakeRW.Cookies) != 1 {
		t.Errorf("len(Cookies) = %v, want 1", len(fakeRW.Cookies))
	}
	if got, want := fakeRW.Cookies[0].String(), "HttpOnly; Secure; SameSite=Strict"; !strings.Contains(got, want) {
		t.Errorf("XSRF cookie got %q, want to contain %q", got, want)
	}
}

func TestMissingCookieInPostRequest(t *testing.T) {
	tests := []struct {
		name       string
		stage      func(it *Interceptor, rw safehttp.ResponseWriter, req *safehttp.IncomingRequest)
		wantStatus safehttp.StatusCode
	}{
		{
			name: "In Before stage",
			stage: func(it *Interceptor, rw safehttp.ResponseWriter, req *safehttp.IncomingRequest) {

				it.Before(rw, req, nil)
			},
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name: "In Commit stage",
			stage: func(it *Interceptor, rw safehttp.ResponseWriter, req *safehttp.IncomingRequest) {
				it.Commit(rw, req, nil, nil)
			},
			wantStatus: safehttp.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader("foo=bar"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			test.stage(&Interceptor{SecretAppKey: "testSecretAppKey"}, fakeRW, req)

			if gotStatus := rr.Code; gotStatus != int(test.wantStatus) {
				t.Errorf("rr.Code: got %v, want %v", gotStatus, test.wantStatus)
			}
		})
	}

}

func TestCommitTokenInResponse(t *testing.T) {
	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	tr := &safehttp.TemplateResponse{}
	i.Commit(fakeRW, req, tr, nil)

	tok, ok := tr.FuncMap["XSRFToken"]
	if !ok {
		t.Fatal(`tr.FuncMap["XSRFToken"] not found`)
	}

	fn, ok := tok.(func() string)
	if !ok {
		t.Fatalf(`tr.FuncMap["XSRFToken"]: got %T, want "func() string"`, fn)
	}
	if got := fn(); got == "" {
		t.Error(`tr.FuncMap["XSRFToken"](): got empty token`, got)
	}

	if want, got := safehttp.StatusOK, rr.Code; got != int(want) {
		t.Errorf("rr.Code: got %v, want %v", got, want)
	}

	if want, got := "", rr.Body.String(); got != want {
		t.Errorf("rr.Body.String(): got %q want %q", got, want)
	}
}

func TestCommitNotTemplateResponse(t *testing.T) {
	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	req := safehttptest.NewRequest(safehttp.MethodGet, "https://foo.com/pizza", nil)

	i := Interceptor{SecretAppKey: "testSecretAppKey"}
	i.Commit(fakeRW, req, safehttp.NoContentResponse{}, nil)

	if want, got := safehttp.StatusOK, rr.Code; got != int(want) {
		t.Errorf("rr.Code: got %v, want %v", got, want)
	}

	if want, got := "", rr.Body.String(); got != want {
		t.Errorf("rr.Body.String(): got %q want %q", got, want)
	}
}
