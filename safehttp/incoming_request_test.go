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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		req := httptest.NewRequest("GET", "/", nil)
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
	req := httptest.NewRequest("GET", "/", nil)
	ir := safehttp.NewIncomingRequest(req)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf(`ir.SetContext(nil): expected panic`)
		}
	}()

	ir.SetContext(nil)
}
