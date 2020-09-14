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

package safehttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

var nilContext context.Context

func TestIncomingRequestCookie(t *testing.T) {
	var tests = []struct {
		name      string
		req       *http.Request
		wantName  string
		wantValue string
	}{
		{
			name: "Basic",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("Cookie", "foo=bar")
				return r
			}(),
			wantName:  "foo",
			wantValue: "bar",
		},
		{
			name: "Multiple cookies with the same name",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Add("Cookie", "foo=bar; foo=xyz")
				r.Header.Add("Cookie", "foo=pizza")
				return r
			}(),
			wantName:  "foo",
			wantValue: "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := safehttp.NewIncomingRequest(tt.req)
			c, err := ir.Cookie(tt.wantName)
			if err != nil {
				t.Errorf(`ir.Cookie(tt.wantName) got: %v want: nil`, err)
			}

			if got := c.Name(); got != tt.wantName {
				t.Errorf("c.Name() got: %v want: %v", got, tt.wantName)
			}

			if got := c.Value(); got != tt.wantValue {
				t.Errorf(`c.Value() got: %v want: %v`, got, tt.wantValue)
			}
		})
	}
}

func TestIncomingRequestCookieNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ir := safehttp.NewIncomingRequest(r)
	if _, err := ir.Cookie("foo"); err == nil {
		t.Error(`ir.Cookie("foo") got: nil want: error`)
	}
}

func TestIncomingRequestCookies(t *testing.T) {
	var tests = []struct {
		name       string
		req        *http.Request
		wantNames  []string
		wantValues []string
	}{
		{
			name: "One",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("Cookie", "foo=bar")
				return r
			}(),
			wantNames:  []string{"foo"},
			wantValues: []string{"bar"},
		},
		{
			name: "Multiple",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Add("Cookie", "foo=bar; a=b")
				r.Header.Add("Cookie", "pizza=hawaii")
				return r
			}(),
			wantNames:  []string{"foo", "a", "pizza"},
			wantValues: []string{"bar", "b", "hawaii"},
		},
		{
			name:       "None",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			wantNames:  []string{},
			wantValues: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := safehttp.NewIncomingRequest(tt.req)
			cl := ir.Cookies()

			if got, want := len(cl), len(tt.wantNames); got != want {
				t.Errorf("len(cl) got: %v want: %v", got, want)
			}

			for i, c := range cl {
				if got := c.Name(); got != tt.wantNames[i] {
					t.Errorf("c.Name() got: %v want: %v", got, tt.wantNames[i])
				}

				if got := c.Value(); got != tt.wantValues[i] {
					t.Errorf(`c.Value() got: %v want: %v`, got, tt.wantValues[i])
				}
			}
		})

	}
}

type pizza struct {
	val string
}

type pizzaKey string

func TestRequestSetValidContextWithValue(t *testing.T) {
	tests := []struct {
		name    string
		key     pizzaKey
		wantVal *pizza
		wantOk  bool
	}{
		{
			name:    "Value set for key",
			key:     pizzaKey("1234"),
			wantOk:  true,
			wantVal: &pizza{val: "margeritta"},
		},
		{
			name:    "Value not set for key",
			key:     pizzaKey("5678"),
			wantOk:  false,
			wantVal: nil,
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest(safehttp.MethodGet, "/", nil)
		ir := safehttp.NewIncomingRequest(req)
		ctx := context.WithValue(ir.Context(), pizzaKey("1234"), &pizza{val: "margeritta"})
		ir.SetContext(ctx)

		got, ok := ir.Context().Value(test.key).(*pizza)
		if ok != test.wantOk {
			t.Errorf("type match: got %v, want %v", ok, test.wantOk)
		}
		if diff := cmp.Diff(test.wantVal, got, cmp.AllowUnexported(pizza{})); diff != "" {
			t.Errorf("ir.Context().Value(test.key): mismatch (-want +got): \n%s", diff)
		}
	}
}

func TestRequestSetNilContext(t *testing.T) {
	req := httptest.NewRequest(safehttp.MethodGet, "/", nil)
	ir := safehttp.NewIncomingRequest(req)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf(`ir.SetContext(nil): expected panic`)
		}
	}()

	ir.SetContext(nilContext)
}

func TestIncomingRequestPostForm(t *testing.T) {
	methods := []string{
		safehttp.MethodPost,
		safehttp.MethodPut,
		safehttp.MethodPatch,
	}

	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			r := safehttptest.NewRequest(m, "/", strings.NewReader("a=b"))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			f, err := r.PostForm()
			if err != nil {
				t.Errorf("r.PostForm() got: %v want: nil", err)
			}

			if got, want := f.String("a", ""), "b"; got != want {
				t.Errorf(`f.String("a", "") got: %q want: %q`, got, want)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestIncomingRequestInvalidPostForm(t *testing.T) {
	tests := []struct {
		name string
		req  *safehttp.IncomingRequest
	}{
		{
			name: "GET method",
			req:  safehttptest.NewRequest(safehttp.MethodGet, "/", nil),
		},
		{
			name: "Wrong content type",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/", nil)
				r.Header.Set("Content-Type", "blah/blah")
				return r
			}(),
		},
		{
			// Note that net/http.Request.ParseForm also parses url parameters and
			// the errors that occur are returned.
			name: "Invalid url parameter",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "http://foo.com/asdf?%xx=yy", nil)
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return r
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.req.PostForm(); err == nil {
				t.Error("tt.req.PostForm() got: nil want: error")
			}
		})
	}
}

func TestIncomingRequestMultipartForm(t *testing.T) {
	methods := []string{
		safehttp.MethodPost,
		safehttp.MethodPut,
		safehttp.MethodPatch,
	}

	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			body := "--123\r\n" +
				"Content-Disposition: form-data; name=\"a\"\r\n" +
				"\r\n" +
				"b\r\n" +
				"--123--\r\n"
			r := safehttptest.NewRequest(m, "/", strings.NewReader(body))
			r.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

			f, err := r.MultipartForm(1000)
			if err != nil {
				t.Errorf("r.MultipartForm(1000) got: %v want: nil", err)
			}

			if got, want := f.String("a", ""), "b"; got != want {
				t.Errorf(`f.String("a", "") got: %q want: %q`, got, want)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestIncomingRequestMultipartFormNegativeMemory(t *testing.T) {
	body := "--123\r\n" +
		"Content-Disposition: form-data; name=\"a\"\r\n" +
		"\r\n" +
		"b\r\n" +
		"--123--\r\n"
	r := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

	f, err := r.MultipartForm(-1)
	if err != nil {
		t.Errorf("r.MultipartForm(-1) got: %v want: nil", err)
	}

	if got, want := f.String("a", ""), "b"; got != want {
		t.Errorf(`f.String("a", "") got: %q want: %q`, got, want)
	}

	if err := f.Err(); err != nil {
		t.Errorf("f.Err() got: %v want: nil", err)
	}
}

func TestIncomingRequestInvalidMultipartForm(t *testing.T) {
	tests := []struct {
		name string
		req  *safehttp.IncomingRequest
	}{
		{
			name: "GET method",
			req:  safehttptest.NewRequest(safehttp.MethodGet, "/", nil),
		},
		{
			name: "Wrong content type",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "/", nil)
				r.Header.Set("Content-Type", "blah/blah")
				return r
			}(),
		},
		{
			// Note that net/http.Request.ParseMultipartForm also parses url parameters
			// and the errors that occur are returned.
			name: "Invalid url parameter",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "http://foo.com/asdf?%xx=yy", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				return r
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.req.MultipartForm(1000)
			if err == nil {
				t.Error("tt.req.ir.MultipartForm(1000) got: nil want: error")
			}
		})
	}
}

func TestIncomingRequestMultipartFileUpload(t *testing.T) {
	body := "--123\r\n" +
		"Content-Disposition: form-data; name=\"file\"; filename=\"myfile\"\r\n" +
		"\r\n" +
		"file content\r\n" +
		"--123--\r\n"
	r := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

	f, err := r.MultipartForm(1024)
	if err != nil {
		t.Errorf("r.MultipartForm(1024): got err %v", err)
	}

	fhs := f.File("file")
	if fhs == nil {
		t.Error(`f.File("file"): got nil, want file header`)
	}
	defer f.RemoveFiles()

	file, err := fhs[0].Open()
	if err != nil {
		t.Fatalf("fhs[0].Open(): got err %v, want nil", err)
	}

	content := make([]byte, 12)
	file.Read(content)
	if want, got := "file content", string(content); want != got {
		t.Errorf("file.Read(content): got %s, want %s", got, want)
	}
}

func TestIncomingRequestMultipartFormAndFileUpload(t *testing.T) {
	body := "--123\r\n" +
		"Content-Disposition: form-data; name=\"key\"\r\n" +
		"\r\n" +
		"12\r\n" +
		"--123\r\n" +
		"Content-Disposition: form-data; name=\"file\"; filename=\"myfile\"\r\n" +
		"\r\n" +
		"file content\r\n" +
		"--123--\r\n"
	r := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

	f, err := r.MultipartForm(1024)
	if err != nil {
		t.Errorf("r.MultipartForm(1024): got err %v", err)
	}

	if want, got := int64(12), f.Int64("key", 0); want != got {
		t.Errorf(`f.Int64("key", 0): got %d, want %d`, got, want)
	}
	if err := f.Err(); err != nil {
		t.Errorf("f.Err(): got err %v", err)
	}

	fhs := f.File("file")
	if fhs == nil {
		t.Error(`f.File("file"): got nil, want file header`)
	}
	defer f.RemoveFiles()

	file, err := fhs[0].Open()
	if err != nil {
		t.Fatalf("fhs[0].Open(): got err %v, want nil", err)
	}

	content := make([]byte, 12)
	file.Read(content)
	if want, got := "file content", string(content); want != got {
		t.Errorf("file.Read(content): got %s, want %s", got, want)
	}
}

func TestIncomingRequestFileUploadMissingContent(t *testing.T) {
	body := "--123\r\n" +
		"Content-Disposition: form-data; name=\"file\"; filename=\"myfile\"\r\n" +
		"\r\n" +
		"--123--\r\n"
	r := safehttptest.NewRequest(safehttp.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)

	f, err := r.MultipartForm(1024)
	if err != nil {
		t.Errorf("r.MultipartForm(1024): got err %v", err)
	}

	fhs := f.File("file")
	if fhs == nil {
		t.Error(`f.File("file"): got nil, want file header`)
	}
	defer f.RemoveFiles()

	file, err := fhs[0].Open()
	if err != nil {
		t.Fatalf("fhs[0].Open(): got err %v, want nil", err)
	}

	content := make([]byte, 0)
	file.Read(content)
	if want, got := "", string(content); want != got {
		t.Errorf("file.Read(content): got %s, want %s", got, want)
	}
}
