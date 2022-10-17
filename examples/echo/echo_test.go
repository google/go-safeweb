// Copyright 2022 Google LLC
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

package main

import (
	"fmt"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-safeweb/safehttp"
)

func TestEcho(t *testing.T) {
	tests := []struct {
		name    string
		req     string
		want    string
		wantErr bool
	}{
		{
			name:    "no error",
			req:     "?message=<h1>h4x0r</h1>",
			want:    "&lt;h1&gt;h4x0r&lt;/h1&gt;",
			wantErr: false,
		},
		{
			name:    "empty message",
			req:     "?message=",
			want:    "",
			wantErr: true,
		}, {
			name:    "invalid query parameters",
			req:     "?message=;something;",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		mb := safehttp.NewServeMuxConfig(nil)
		mux := mb.Mux()
		mux.Handle("/", safehttp.MethodGet, safehttp.HandlerFunc(echo))

		req := httptest.NewRequest(safehttp.MethodGet, fmt.Sprintf("http://foo.com/%s", tt.req), nil)
		rw := httptest.NewRecorder()
		mux.ServeHTTP(rw, req)

		if rw.Code != int(safehttp.StatusOK) && !tt.wantErr {
			t.Errorf("echo() status = %v, wantErr = %v", rw.Code, tt.wantErr)
		}

		if body := rw.Body.String(); !tt.wantErr && body != tt.want {
			t.Errorf("body got: %q want: %q", body, tt.want)
		}
	}
}

func TestUptime(t *testing.T) {
	tests := []struct {
		name    string
		req     string
		wantErr bool
	}{
		{
			name:    "no error",
			req:     "",
			wantErr: false,
		},
		{
			name:    "invalid query parameters",
			req:     "?message=;something;",
			wantErr: true,
		},
	}

	rgx := regexp.MustCompile("^<h1>Uptime: (.*)</h1>$")

	for _, tt := range tests {
		mb := safehttp.NewServeMuxConfig(nil)
		mux := mb.Mux()
		start = time.Date(1991, time.September, 17, 00, 00, 00, 00, time.UTC)
		mux.Handle("/", safehttp.MethodGet, safehttp.HandlerFunc(uptime))

		req := httptest.NewRequest(safehttp.MethodGet, fmt.Sprintf("http://foo.com/%s", tt.req), nil)
		rw := httptest.NewRecorder()
		mux.ServeHTTP(rw, req)

		if rw.Code != int(safehttp.StatusOK) && !tt.wantErr {
			t.Errorf("uptime() status = %v, wantErr = %v", rw.Code, tt.wantErr)
		}

		if !tt.wantErr {
			matched := rgx.Match(rw.Body.Bytes())
			if !matched {
				t.Errorf("body got: %q want: %q", rw.Body.String(), "<h1>Uptime: X</h1>")
			}
		}
	}
}
