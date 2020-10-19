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
	"math"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml"
)

func TestResponseWriterSetCookie(t *testing.T) {
	testRw := safehttptest.NewTestResponseWriter(&strings.Builder{})
	rw := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, testRw, nil)

	c := safehttp.NewCookie("foo", "bar")
	err := rw.SetCookie(c)
	if err != nil {
		t.Errorf("rw.SetCookie(c) got: %v want: nil", err)
	}

	wantHeaders := map[string][]string{
		"Set-Cookie": {"foo=bar; HttpOnly; Secure; SameSite=Lax"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(testRw.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}
}

func TestResponseWriterSetInvalidCookie(t *testing.T) {
	testRw := safehttptest.NewTestResponseWriter(&strings.Builder{})
	rw := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, testRw, nil)

	c := safehttp.NewCookie("f=oo", "bar")
	err := rw.SetCookie(c)
	if err == nil {
		t.Error("rw.SetCookie(c) got: nil want: error")
	}
}

func TestResponseWriterWriteTwicePanic(t *testing.T) {
	tests := []struct {
		name  string
		write func(w *safehttp.ResponseWriter)
	}{
		{
			name: "Call Write twice",
			write: func(w *safehttp.ResponseWriter) {
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Call NoContent twice",
			write: func(w *safehttp.ResponseWriter) {
				w.NoContent()
				w.NoContent()
			},
		},
		{
			name: "Call WriteError twice",
			write: func(w *safehttp.ResponseWriter) {
				w.WriteError(safehttp.StatusInternalServerError)
				w.WriteError(safehttp.StatusInternalServerError)
			},
		},
		{
			name: "Call Redirect twice",
			write: func(w *safehttp.ResponseWriter) {
				ir := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
				w.Redirect(ir, "/asdf", safehttp.StatusMovedPermanently)
				w.Redirect(ir, "/asdf", safehttp.StatusMovedPermanently)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, safehttptest.NewTestResponseWriter(&strings.Builder{}), nil)
			defer func() {
				if r := recover(); r != nil {
					return
				}
				t.Errorf("tt.write(w) expected panic")
			}()
			tt.write(w)
		})
	}
}

func TestResponseWriterUnsafeResponse(t *testing.T) {
	tests := []struct {
		name  string
		write func(w *safehttp.ResponseWriter)
	}{
		{
			name: "Unsafe HTML Response",
			write: func(w *safehttp.ResponseWriter) {
				w.Write("<h1>Hello World!</h1>")
			},
		},
		{
			name: "Invalid JSON Response",
			write: func(w *safehttp.ResponseWriter) {
				safehttp.WriteJSON(w, math.Inf(1))
			},
		},
		{
			name: "Unsafe Template Response",
			write: func(w *safehttp.ResponseWriter) {
				safehttp.ExecuteTemplate(w, template.
					Must(template.New("name").
						Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, safehttptest.NewTestResponseWriter(&strings.Builder{}), nil)
			defer func() {
				if r := recover(); r != nil {
					return
				}
				t.Errorf("tt.write(w) expected panic")
			}()
			tt.write(w)
		})
	}
}

func TestResponseWriterStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		write      func(w *safehttp.ResponseWriter)
		wantStatus safehttp.StatusCode
	}{
		{
			name: "Default status code",
			write: func(w *safehttp.ResponseWriter) {
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
			wantStatus: safehttp.StatusOK,
		},
		{
			name: "Custom status code",
			write: func(w *safehttp.ResponseWriter) {
				w.SetCode(safehttp.StatusPartialContent)
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
			wantStatus: safehttp.StatusPartialContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRw := safehttptest.NewTestResponseWriter(&strings.Builder{})
			w := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, testRw, nil)
			tt.write(w)
			if gotStatus := testRw.Status(); gotStatus != tt.wantStatus {
				t.Errorf("tt.write(w): got %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}

func TestResponseWriterInvalidStatusCode(t *testing.T) {
	tests := []struct {
		name  string
		write func(w *safehttp.ResponseWriter)
	}{
		{
			name: "Status code smaller than 100",
			write: func(w *safehttp.ResponseWriter) {
				w.SetCode(50)
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Status code bigger than 599",
			write: func(w *safehttp.ResponseWriter) {
				w.SetCode(1000)
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Status code set after write",
			write: func(w *safehttp.ResponseWriter) {
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
				w.SetCode(safehttp.StatusPartialContent)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, safehttptest.NewTestResponseWriter(&strings.Builder{}), nil)
			defer func() {
				if r := recover(); r != nil {
					return
				}
				t.Errorf("tt.write(w) expected panic")
			}()
			tt.write(w)
		})
	}
}
