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

// +build go1.16

package safehttp_test

import (
	"embed"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

//go:embed testdata
var testEmbeddedFS embed.FS

func TestFileServerEmbed(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantCode safehttp.StatusCode
		wantCT   string
		wantBody string
	}{
		{
			name:     "missing file",
			path:     "failure",
			wantCode: 404,
			wantCT:   "text/plain; charset=utf-8",
			wantBody: "Not Found\n",
		},
		{
			name:     "embedded file",
			path:     "testdata/embed.html",
			wantCode: 200,
			wantCT:   "text/html; charset=utf-8",
			wantBody: "<h1> This is a test embedded document </h1>\n",
		},
	}

	mb := &safehttp.ServeMuxConfig{}
	mb.Handle("/", safehttp.MethodGet, safehttp.FileServerEmbed(testEmbeddedFS))
	m := mb.Mux()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			rr := safehttptest.NewTestResponseWriter(&b)

			req := httptest.NewRequest(safehttp.MethodGet, "https://test.science/"+tt.path, nil)
			m.ServeHTTP(rr, req)

			if got, want := rr.Status(), tt.wantCode; got != tt.wantCode {
				t.Errorf("status code got: %v want: %v", got, want)
			}
			if got := rr.Header().Get("Content-Type"); tt.wantCT != got {
				t.Errorf("Content-Type: got %q want %q", got, tt.wantCT)
			}
			if diff := cmp.Diff(tt.wantBody, b.String()); diff != "" {
				t.Errorf("Response body diff (-want,+got): \n%s", diff)
			}
		})
	}
}
