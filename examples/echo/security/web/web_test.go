// Copyright 2022 Google LLC
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

package web

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

func TestNewMuxConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *safehttp.ServeMuxConfig
		addr    string
		hasHSTS bool
	}{
		{
			name:    "development mux config",
			config:  NewMuxConfigDev(8080),
			addr:    "https://localhost:8080",
			hasHSTS: false,
		},
		{
			name:    "production mux config",
			config:  NewMuxConfig("localhost:8080"),
			addr:    "https://localhost:8080",
			hasHSTS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := tt.config.Mux()
			h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mux.Handle("/spaghetti", safehttp.MethodGet, h)

			req := httptest.NewRequest(safehttp.MethodGet, fmt.Sprintf("%s/spaghetti", tt.addr), nil)
			rw := httptest.NewRecorder()
			mux.ServeHTTP(rw, req)

			hasHSTSHeader := rw.Header().Get("Strict-Transport-Security") != ""
			if tt.hasHSTS && !hasHSTSHeader {
				t.Errorf("expected \"Strict-Transport-Security\" header since HTST is enabled")
			}
			if !tt.hasHSTS && hasHSTSHeader {
				t.Errorf("unexpected \"Strict-Transport-Security\" header since HTST is disabled")
			}
		})
	}
}
