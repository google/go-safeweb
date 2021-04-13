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
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestFileServer(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-safehttp-test")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := ioutil.WriteFile(tmpDir+"/foo.html", []byte("<h1>Hello world</h1>"), 0644); err != nil {
		t.Fatalf("ioutil.WriteFile: %v", err)
	}

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
			name:     "file",
			path:     "foo.html",
			wantCode: 200,
			wantCT:   "text/html; charset=utf-8",
			wantBody: "<h1>Hello world</h1>",
		},
	}

	mb := &safehttp.ServeMuxConfig{}
	mb.Handle("/", safehttp.MethodGet, safehttp.FileServer(tmpDir))
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

// TODO(kele): Add tests including interceptors once we have
// https://github.com/google/go-safeweb/issues/261.
