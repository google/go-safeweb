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
	"testing"

	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

func TestNewMuxConfig(t *testing.T) {
	cf, addr := newServeMuxConfig()
	mux := cf.Mux()
	h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/", safehttp.MethodGet, h)

	req := httptest.NewRequest(safehttp.MethodGet, fmt.Sprintf("http://%s/", addr), nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	body := rw.Body.String()
	if want := "&lt;h1&gt;Hello World!&lt;/h1&gt;"; body != want {
		t.Errorf("body got: %q want: %q", body, want)
	}
}

func Test_loadTemplate(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name:    "existing template, no error",
			src:     "safe.html",
			wantErr: false,
		},
		{
			name:    "missing template, error",
			src:     "not_existing.html",
			wantErr: true,
		},
		{
			name:    "invalid source name, error",
			src:     "../../hidden.html",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("unexpected panic: %v", r)
				}
			}()
			_, err := loadTemplate(tt.src)
			if err != nil && !tt.wantErr {
				t.Errorf("loadTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestHandleTemplate(t *testing.T) {
	mux := safehttp.NewServeMuxConfig(nil).Mux()
	safeTmpl := template.Must(template.New("standard").Parse(`<h1>Hi there!</h1>`))
	mux.Handle("/", safehttp.MethodGet, safehttp.HandlerFunc(handleTemplate(safeTmpl)))

	req := httptest.NewRequest(safehttp.MethodGet, "/spaghetti", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	if body, want := rw.Body.String(), "<h1>Hi there!</h1>"; body != want {
		t.Errorf("handleTemplate() got %q, want %q", body, want)
	}
}
