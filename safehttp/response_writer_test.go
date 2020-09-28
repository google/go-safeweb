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
			name: "Call WriteJSON twice",
			write: func(w *safehttp.ResponseWriter) {
				obj := struct{ Field string }{Field: "myField"}
				w.WriteJSON(obj)
				w.WriteJSON(obj)
			},
		},
		{
			name: "Call Write then WriteJSON",
			write: func(w *safehttp.ResponseWriter) {
				obj := struct{ Field string }{Field: "myField"}
				w.Write(obj)
				w.WriteJSON(obj)
			},
		},
		{
			name: "Call WriteTemplate twice",
			write: func(w *safehttp.ResponseWriter) {
				w.WriteTemplate(template.Must(template.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
				w.WriteTemplate(template.Must(template.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
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
				if r := recover(); r == nil {
					t.Errorf("tt.write(w) expected panic")
				}
			}()
			tt.write(w)
		})
	}
}

func TestResponseWriterStatusCode(t *testing.T) {
	tests := []struct {
		name  string
		want  safehttp.StatusCode
		write func(w *safehttp.ResponseWriter)
	}{
		{
			name: "Status code not explicitly set",
			want: safehttp.StatusOK,
			write: func(w *safehttp.ResponseWriter) {
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Custom status code before calling write",
			want: safehttp.StatusCreated,
			write: func(w *safehttp.ResponseWriter) {
				w.Code = safehttp.StatusCreated
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Custom status code after calling write",
			want: safehttp.StatusOK,
			write: func(w *safehttp.ResponseWriter) {
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
				w.Code = safehttp.StatusCreated
			},
		},

		{
			name: "No content status code",
			want: safehttp.StatusNoContent,
			write: func(w *safehttp.ResponseWriter) {
				w.NoContent()
			},
		},
		{
			name: "Error status code",
			want: safehttp.StatusInternalServerError,
			write: func(w *safehttp.ResponseWriter) {
				w.WriteError(safehttp.StatusInternalServerError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRw := safehttptest.NewTestResponseWriter(&strings.Builder{})
			w := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, testRw, nil)

			tt.write(w)

			if got := testRw.Status(); tt.want != got {
				t.Errorf("testRw.Status(): got %v want %v", got, tt.want)
			}
		})
	}
}

func TestResponseWriterStatusCodePanic(t *testing.T) {
	tests := []struct {
		name  string
		write func(w *safehttp.ResponseWriter)
	}{
		{
			name: "Manual redirect status code set",
			write: func(w *safehttp.ResponseWriter) {
				w.Code = safehttp.StatusPermanentRedirect
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Manual error status code set",
			write: func(w *safehttp.ResponseWriter) {
				w.Code = safehttp.StatusBadRequest
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
		{
			name: "Invalid status code set",
			write: func(w *safehttp.ResponseWriter) {
				w.Code = safehttp.StatusCode(50)
				w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := safehttp.NewResponseWriter(safehttp.DefaultDispatcher{}, safehttptest.NewTestResponseWriter(&strings.Builder{}), nil)
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("tt.write(w) expected panic")
				}
			}()
			tt.write(w)
		})
	}
}
