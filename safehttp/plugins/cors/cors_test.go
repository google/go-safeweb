// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cors_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/cors"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestRequest(t *testing.T) {
	tests := []struct {
		name             string
		req              *safehttp.IncomingRequest
		allowCredentials bool
		exposedHeaders   []string
		want             map[string][]string
	}{
		{
			name: "Basic GET",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodGet, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			want: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.com"},
				"Vary":                        {"Origin"},
			},
		},
		{
			name: "Basic PUT",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			want: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.com"},
				"Vary":                        {"Origin"},
			},
		},
		{
			name: "Basic POST",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPost, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			want: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.com"},
				"Vary":                        {"Origin"},
			},
		},
		{
			name: "No Origin header",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			want: map[string][]string{},
		},
		{
			name: "AllowCredentials but no cookies",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			allowCredentials: true,
			want: map[string][]string{
				"Access-Control-Allow-Origin": {"https://foo.com"},
				"Vary":                        {"Origin"},
			},
		},
		{
			name: "AllowCredentials with cookies",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Cookie", "a=b")
				return r
			}(),
			allowCredentials: true,
			want: map[string][]string{
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Allow-Origin":      {"https://foo.com"},
				"Vary":                             {"Origin"},
			},
		},
		{
			name: "Expose one header",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			exposedHeaders: []string{"Aaaa"},
			want: map[string][]string{
				"Access-Control-Expose-Headers": {"Aaaa"},
				"Access-Control-Allow-Origin":   {"https://foo.com"},
				"Vary":                          {"Origin"},
			},
		},
		{
			name: "Expose multiple headers",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("X-Cors", "1")
				r.Header.Set("Content-Type", "application/json")
				return r
			}(),
			exposedHeaders: []string{"Aaaa", "Bbbb", "Cccc"},
			want: map[string][]string{
				"Access-Control-Expose-Headers": {"Aaaa, Bbbb, Cccc"},
				"Access-Control-Allow-Origin":   {"https://foo.com"},
				"Vary":                          {"Origin"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			it := cors.Default("https://foo.com")
			it.AllowCredentials = tt.allowCredentials
			it.ExposedHeaders = tt.exposedHeaders
			it.Before(fakeRW, tt.req, nil)

			if diff := cmp.Diff(tt.want, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
			if got := rr.Body.String(); got != "" {
				t.Errorf(`rr.Body.String() got: %q want: ""`, got)
			}
		})
	}
}

func TestVaryHeaderAppending(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
	req.Header.Set("Origin", "https://foo.com")
	req.Header.Set("X-Cors", "1")
	req.Header.Set("Content-Type", "application/json")

	fakeRW, rr := safehttptest.NewFakeResponseWriter()
	rr.Header().Set("Vary", "a")

	it := cors.Default("https://foo.com")
	it.Before(fakeRW, req, nil)

	wantHeaders := map[string][]string{
		"Access-Control-Allow-Origin": {"https://foo.com"},
		"Vary":                        {"a, Origin"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
	if got := rr.Body.String(); got != "" {
		t.Errorf(`rr.Body.String() got: %q want: ""`, got)
	}
}

func TestHeadRequest(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodHead, "http://bar.com", nil)
	req.Header.Set("Origin", "https://foo.com")
	req.Header.Set("X-Cors", "1")
	req.Header.Set("Content-Type", "application/json")

	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	it := cors.Default("https://foo.com")
	it.Before(fakeRW, req, nil)

	if got, want := rr.Code, int(safehttp.StatusMethodNotAllowed); got != want {
		t.Errorf("rr.Code got: %v want: %v", got, want)
	}
	wantHeaders := map[string][]string{}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
}

func TestInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *safehttp.IncomingRequest
	}{
		{
			name: "No X-Cors: 1, but Sec-Fetch-Mode: cors",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Sec-Fetch-Mode", "cors")
				return r
			}(),
		},
		{
			name: "No X-Cors: 1",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				return r
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			it := cors.Default("https://foo.com")
			it.Before(fakeRW, tt.req, nil)

			if want := safehttp.StatusPreconditionFailed; rr.Code != int(want) {
				t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
			}
			wantHeaders := map[string][]string{}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestDisallowedContentTypes(t *testing.T) {
	contentTypes := []string{
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
		"",
	}

	for _, ct := range contentTypes {
		t.Run(ct, func(t *testing.T) {
			req := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com/asdf", nil)
			req.Header.Set("Origin", "https://foo.com")
			req.Header.Set("X-Cors", "1")
			if ct != "" {
				req.Header.Set("Content-Type", ct)
			}

			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			it := cors.Default("https://foo.com")
			it.Before(fakeRW, req, nil)

			if want := safehttp.StatusUnsupportedMediaType; rr.Code != int(want) {
				t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
			}
			wantHeaders := map[string][]string{}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDisallowedOrigin(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodPut, "http://bar.com/asdf", nil)
	req.Header.Set("Origin", "https://pizza.com")

	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	it := cors.Default("https://foo.com")
	it.Before(fakeRW, req, nil)

	if want := safehttp.StatusForbidden; rr.Code != int(want) {
		t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
	}
	wantHeaders := map[string][]string{}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
}

func TestPreflight(t *testing.T) {
	tests := []struct {
		name           string
		req            *safehttp.IncomingRequest
		allowedHeaders []string
		maxAge         int
		wantHeaders    map[string][]string
	}{
		{
			name: "Basic",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Access-Control-Request-Method", safehttp.MethodPut)
				return r
			}(),
			wantHeaders: map[string][]string{
				"Access-Control-Allow-Methods": {"PUT"},
				"Access-Control-Allow-Origin":  {"https://foo.com"},
				"Access-Control-Max-Age":       {"5"},
				"Vary":                         {"Origin"},
			},
		},
		{
			name: "Request X-Cors header",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Access-Control-Request-Method", safehttp.MethodPut)
				r.Header.Set("Access-Control-Request-Headers", "X-Cors")
				return r
			}(),
			wantHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"X-Cors"},
				"Access-Control-Allow-Methods": {"PUT"},
				"Access-Control-Allow-Origin":  {"https://foo.com"},
				"Access-Control-Max-Age":       {"5"},
				"Vary":                         {"Origin"},
			},
		},
		{
			name: "Request custom header",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Access-Control-Request-Method", safehttp.MethodPut)
				r.Header.Set("Access-Control-Request-Headers", "Aaaa")
				return r
			}(),
			allowedHeaders: []string{"Aaaa"},
			wantHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"Aaaa"},
				"Access-Control-Allow-Methods": {"PUT"},
				"Access-Control-Allow-Origin":  {"https://foo.com"},
				"Access-Control-Max-Age":       {"5"},
				"Vary":                         {"Origin"},
			},
		},
		{
			name: "Request multiple headers",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Access-Control-Request-Method", safehttp.MethodPut)
				r.Header.Set("Access-Control-Request-Headers", "X-Cors, Aaaa")
				return r
			}(),
			allowedHeaders: []string{"Aaaa"},
			wantHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"X-Cors, Aaaa"},
				"Access-Control-Allow-Methods": {"PUT"},
				"Access-Control-Allow-Origin":  {"https://foo.com"},
				"Access-Control-Max-Age":       {"5"},
				"Vary":                         {"Origin"},
			},
		},
		{
			name: "Request headers test canonicalization",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Access-Control-Request-Method", safehttp.MethodPut)
				r.Header.Set("Access-Control-Request-Headers", "x-coRS, aaAA")
				return r
			}(),
			allowedHeaders: []string{"AAaa"},
			wantHeaders: map[string][]string{
				"Access-Control-Allow-Headers": {"x-coRS, aaAA"},
				"Access-Control-Allow-Methods": {"PUT"},
				"Access-Control-Allow-Origin":  {"https://foo.com"},
				"Access-Control-Max-Age":       {"5"},
				"Vary":                         {"Origin"},
			},
		},
		{
			name: "Custom Max age",
			req: func() *safehttp.IncomingRequest {
				r := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
				r.Header.Set("Origin", "https://foo.com")
				r.Header.Set("Access-Control-Request-Method", safehttp.MethodPut)
				return r
			}(),
			maxAge: 3600,
			wantHeaders: map[string][]string{
				"Access-Control-Allow-Methods": {"PUT"},
				"Access-Control-Allow-Origin":  {"https://foo.com"},
				"Access-Control-Max-Age":       {"3600"},
				"Vary":                         {"Origin"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			it := cors.Default("https://foo.com")
			it.MaxAge = tt.maxAge
			it.SetAllowedHeaders(tt.allowedHeaders...)
			it.Before(fakeRW, tt.req, nil)

			if rr.Code != int(safehttp.StatusNoContent) {
				t.Errorf("rr.Code got: %v want: %v", rr.Code, safehttp.StatusNoContent)
			}
			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
			if got := rr.Body.String(); got != "" {
				t.Errorf(`rr.Body.String() got: %q want: ""`, got)
			}
		})
	}
}

func TestInvalidAccessControlRequestHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers string
	}{
		{
			name:    "B is not allowed",
			headers: "B",
		},
		{
			name:    "One in list is not allowed",
			headers: "X-Cors, B",
		},
		{
			name:    "Empty at the end",
			headers: "X-Cors, ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
			rh := req.Header
			rh.Set("Origin", "https://foo.com")
			rh.Set("Access-Control-Request-Method", safehttp.MethodPut)
			rh.Set("Access-Control-Request-Headers", tt.headers)

			fakeRW, rr := safehttptest.NewFakeResponseWriter()

			it := cors.Default("https://foo.com")
			it.Before(fakeRW, req, nil)

			if want := safehttp.StatusForbidden; rr.Code != int(want) {
				t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
			}
			wantHeaders := map[string][]string{}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEmptyAccessControlRequestMethod(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
	rh := req.Header
	rh.Set("Origin", "https://foo.com")

	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	it := cors.Default("https://foo.com")
	it.Before(fakeRW, req, nil)

	if want := safehttp.StatusForbidden; rr.Code != int(want) {
		t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
	}
	wantHeaders := map[string][]string{}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
}

func TestAccessControlRequestMethodHead(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
	req.Header.Set("Origin", "https://foo.com")
	req.Header.Set("Access-Control-Request-Method", safehttp.MethodHead)

	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	it := cors.Default("https://foo.com")
	it.Before(fakeRW, req, nil)

	if want := safehttp.StatusForbidden; rr.Code != int(want) {
		t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
	}
	wantHeaders := map[string][]string{}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
}

func TestPreflightEmptyOrigin(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodOptions, "http://bar.com/asdf", nil)
	req.Header.Set("Access-Control-Request-Method", safehttp.MethodHead)

	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	it := cors.Default("https://foo.com")
	it.Before(fakeRW, req, nil)

	if want := safehttp.StatusForbidden; rr.Code != int(want) {
		t.Errorf("rr.Code got: %v want: %v", rr.Code, want)
	}
	wantHeaders := map[string][]string{}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
}
