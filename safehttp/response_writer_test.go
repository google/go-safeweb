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
	rr := newResponseRecorder(&strings.Builder{})
<<<<<<< HEAD
	rw := safehttp.NewResponseWriter(testDispatcher{}, rr)
=======
	rw := safehttp.NewResponseWriter(testDispatcher{}, rr, nil, nil)
>>>>>>> Added the Commit phase.

	c := safehttp.NewCookie("foo", "bar")
	err := rw.SetCookie(c)
	if err != nil {
		t.Errorf("rw.SetCookie(c) got: %v want: nil", err)
	}

	wantHeaders := map[string][]string{
		"Set-Cookie": {"foo=bar; HttpOnly; Secure; SameSite=Lax"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.header)); diff != "" {
		t.Errorf("rr.header mismatch (-want +got):\n%s", diff)
	}
}

func TestResponseWriterSetInvalidCookie(t *testing.T) {
	rr := newResponseRecorder(&strings.Builder{})
	rw := safehttp.NewResponseWriter(testDispatcher{}, rr, nil, nil)

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
			w := safehttp.NewResponseWriter(testDispatcher{}, newResponseRecorder(&strings.Builder{}), nil, nil)
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("tt.write(w) expected panic")
				}
			}()
			tt.write(w)
		})
	}
}
