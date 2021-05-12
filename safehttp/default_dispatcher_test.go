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
	"html/template"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
	safetemplate "github.com/google/safehtml/template"
)

func TestDefaultDispatcherValidResponse(t *testing.T) {
	tests := []struct {
		name        string
		write       func(w http.ResponseWriter) error
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "Safe HTML Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				return d.Write(w, safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name: "Safe HTML Template Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				t := safehttp.Template(safetemplate.
					Must(safetemplate.New("name").
						Parse("<h1>{{ . }}</h1>")))
				var data interface{}
				data = "This is an actual heading, though."
				return d.Write(w, &safehttp.TemplateResponse{Template: t, Data: data})
			},
			wantBody: "<h1>This is an actual heading, though.</h1>",
		},
		{
			name: "Named Safe HTML Template Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				t := safehttp.Template(
					safetemplate.Must(
						safetemplate.Must(safetemplate.New("name").Parse("<h1>{{ . }}</h1>")).
							New("associated").Parse("<h2>{{.}}</h2>")))
				var data interface{}
				data = "This is an actual heading, though."
				return d.Write(w, &safehttp.TemplateResponse{t, "associated", data, nil})
			},
			wantBody: "<h2>This is an actual heading, though.</h2>",
		},
		{
			name: "Safe HTML Template Response with Token",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				defaultFunc := func() string { panic("this function should never be called") }
				t := safehttp.Template(safetemplate.
					Must(safetemplate.New("name").
						Funcs(map[string]interface{}{"Token": defaultFunc}).
						Parse(`<form><input type="hidden" name="token" value="{{Token}}">{{.}}</form>`)))
				var data interface{}
				data = "Content"
				fm := map[string]interface{}{
					"Token": func() string { return "Token-secret" },
				}
				return d.Write(w, &safehttp.TemplateResponse{Template: t, Data: data, FuncMap: fm})
			},
			wantBody: `<form><input type="hidden" name="token" value="Token-secret">Content</form>`,
		},
		{
			name: "Safe HTML Template Response with  Nonce",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				defaultFunc := func() string { panic("this function should never be called") }
				t := safehttp.Template(safetemplate.
					Must(safetemplate.New("name").
						Funcs(map[string]interface{}{"Nonce": defaultFunc}).
						Parse(`<script nonce="{{Nonce}}" type="application/javascript">alert("script")</script><h1>{{.}}</h1>`)))
				var data interface{}
				data = "Content"
				fm := map[string]interface{}{
					"Nonce": func() string { return "Nonce-secret" },
				}
				return d.Write(w, &safehttp.TemplateResponse{Template: t, Data: data, FuncMap: fm})
			},
			wantBody: `<script nonce="Nonce-secret" type="application/javascript">alert("script")</script><h1>Content</h1>`,
		},
		{
			name: "Valid JSON Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				data := struct {
					Field string `json:"field"`
				}{Field: "myField"}
				return d.Write(w, safehttp.JSONResponse{data})
			},
			wantBody: ")]}',\n{\"field\":\"myField\"}\n",
		},
		{
			name: "Redirect Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				req := httptest.NewRequest("GET", "/path", nil)
				r := safehttp.NewIncomingRequest(req)
				return d.Write(w, safehttp.RedirectResponse{Request: r, Location: "/anotherpath", Code: safehttp.StatusFound})
			},
			wantHeaders: map[string][]string{"Location": {"/anotherpath"}},
			wantStatus:  safehttp.StatusFound,
			wantBody:    "<a href=\"/anotherpath\">Found</a>.\n\n",
		},
		{
			name: "No Content Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				return d.Write(w, safehttp.NoContentResponse{})
			},
			wantBody:   "",
			wantStatus: safehttp.StatusNoContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			err := tt.write(rw)

			if err != nil {
				t.Errorf("tt.write(rw): got error %v, want nil", err)
			}

			if gotBody := rw.Body.String(); tt.wantBody != gotBody {
				t.Errorf("response body: got %q, want %q", gotBody, tt.wantBody)
			}

			for k, want := range tt.wantHeaders {
				got := rw.Header().Values(k)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("response header %q: -want +got %s", k, diff)
				}
			}

			wantStatus := tt.wantStatus
			if wantStatus == 0 {
				wantStatus = 200
			}

			if got := rw.Code; got != int(wantStatus) {
				t.Errorf("Status: got %d, want %d", got, wantStatus)
			}
		})
	}
}

func TestDefaultDispatcherInvalidResponse(t *testing.T) {
	tests := []struct {
		name  string
		write func(w http.ResponseWriter) error
		want  string
	}{
		{
			name: "Unsafe HTML Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				return d.Write(w, "<h1>Hello World!</h1>")
			},
			want: "",
		},
		{
			name: "Unsafe Template Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				t := safehttp.Template(template.
					Must(template.New("name").
						Parse("<h1>{{ . }}</h1>")))
				var data interface{}
				data = "This is an actual heading, though."
				return d.Write(w, safehttp.TemplateResponse{Template: t, Data: data})
			},
			want: "",
		},
		{
			name: "Invalid JSON Response",
			write: func(w http.ResponseWriter) error {
				d := &safehttp.DefaultDispatcher{}
				return d.Write(w, safehttp.JSONResponse{math.Inf(1)})
			},
			want: ")]}',\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := httptest.NewRecorder()

			if err := tt.write(rw); err == nil {
				t.Error("tt.write(rw): got nil, want error")
			}

			if want, got := tt.want, rw.Body.String(); want != got {
				t.Errorf("response body: got %q, want %q", got, want)
			}
		})
	}
}
