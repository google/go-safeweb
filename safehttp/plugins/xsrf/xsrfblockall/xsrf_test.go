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

package xsrfblockall

import (
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestStateChanging(t *testing.T) {
	test := []struct {
		name       string
		host       string
		method     string
		wantStatus safehttp.StatusCode
	}{
		{
			name:       "POST request",
			host:       "foo.com",
			method:     safehttp.MethodPost,
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name:       "PUT request",
			host:       "foo.com",
			method:     safehttp.MethodPut,
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name:       "DELETE request",
			host:       "foo.com",
			method:     safehttp.MethodDelete,
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name:       "PATCH request",
			host:       "foo.com",
			method:     safehttp.MethodPatch,
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name:       "TRACE request",
			host:       "foo.com",
			method:     safehttp.MethodTrace,
			wantStatus: safehttp.StatusForbidden,
		},
		{
			name:       "CONNECT request",
			host:       "foo.com",
			method:     safehttp.MethodConnect,
			wantStatus: safehttp.StatusForbidden,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(test.method, "https://foo.com/", nil)

			i := Interceptor{}
			i.Before(fakeRW, req, nil)

			if got := rr.Code; got != int(test.wantStatus) {
				t.Errorf("rr.Code: got %v, want %v", got, test.wantStatus)
			}
		})
	}
}

func TestNonStateChanging(t *testing.T) {
	test := []struct {
		name       string
		host       string
		method     string
		wantStatus safehttp.StatusCode
	}{
		{
			name:       "GET request",
			host:       "foo.com",
			method:     safehttp.MethodGet,
			wantStatus: safehttp.StatusOK,
		},
		{
			name:       "HEAD request",
			host:       "foo.com",
			method:     safehttp.MethodHead,
			wantStatus: safehttp.StatusOK,
		},
		{
			name:       "OPTIONS request",
			host:       "foo.com",
			method:     safehttp.MethodOptions,
			wantStatus: safehttp.StatusOK,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(test.method, "https://foo.com/", nil)

			i := Interceptor{}
			i.Before(fakeRW, req, nil)

			if got := rr.Code; got != int(test.wantStatus) {
				t.Errorf("rr.Code: got %v, want %v", got, test.wantStatus)
			}
		})
	}
}
