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

package hsts_test

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/hsts"
	"github.com/google/safehtml"
)

type testDispatcher struct{}

func (testDispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (testDispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

type responseRecorder struct {
	header http.Header
	writer io.Writer
	status int
}

func newResponseRecorder(w io.Writer) *responseRecorder {
	return &responseRecorder{header: http.Header{}, writer: w, status: http.StatusOK}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

func TestHSTS(t *testing.T) {
	var test = []struct {
		name        string
		plugin      hsts.Plugin
		req         *http.Request
		wantStatus  int
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:       "HTTPS",
			plugin:     hsts.NewPlugin(),
			req:        httptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=63072000; includeSubDomains"},
			},
			wantBody: "",
		},
		{
			name:       "HTTP",
			plugin:     hsts.NewPlugin(),
			req:        httptest.NewRequest("GET", "http://localhost/", nil),
			wantStatus: 301,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Location":     {"https://localhost/"},
			},
			wantBody: "<a href=\"https://localhost/\">Moved Permanently</a>.\n\n",
		},
		{
			name:       "HTTP behind proxy",
			plugin:     hsts.Plugin{BehindProxy: true},
			req:        httptest.NewRequest("GET", "http://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=0; includeSubDomains"},
			},
			wantBody: "",
		},
		{
			name:       "Preload",
			plugin:     hsts.Plugin{Preload: true, DisableIncludeSubDomains: true},
			req:        httptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=0; preload"},
			},
			wantBody: "",
		},
		{
			name:       "Preload and IncludeSubDomains",
			plugin:     hsts.Plugin{Preload: true},
			req:        httptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=0; includeSubDomains; preload"},
			},
			wantBody: "",
		},
		{
			name:       "No preload and no includeSubDomains",
			plugin:     hsts.Plugin{DisableIncludeSubDomains: true},
			req:        httptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=0"},
			},
			wantBody: "",
		},
		{
			name:       "Custom maxage",
			plugin:     hsts.Plugin{MaxAge: 7777},
			req:        httptest.NewRequest("GET", "https://localhost/", nil),
			wantStatus: 200,
			wantHeaders: map[string][]string{
				"Strict-Transport-Security": {"max-age=7777; includeSubDomains"},
			},
			wantBody: "",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			m := safehttp.NewMachinery(tt.plugin.Before, &testDispatcher{})

			b := &strings.Builder{}
			rr := newResponseRecorder(b)

			m.HandleRequest(rr, tt.req)

			if rr.status != tt.wantStatus {
				t.Errorf("status code got: %v want: %v", rr.status, tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := b.String(); got != tt.wantBody {
				t.Errorf("response body got: %q want: %q", got, tt.wantBody)
			}
		})
	}
}

func TestStrictTransportSecurityAlreadyImmutable(t *testing.T) {
	p := hsts.NewPlugin()
	m := safehttp.NewMachinery(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		w.Header().MarkImmutable("Strict-Transport-Security")
		return p.Before(w, r)
	}, &testDispatcher{})

	req := httptest.NewRequest("GET", "https://localhost/", nil)

	// TODO(@mattiasgrenfeldt): Change from current to desired behavior once
	// safehttp.ResponseWriter.ServerError has changed signature and been
	// implemented.
	t.Run("Current Behavior", func(t *testing.T) {
		b := &strings.Builder{}
		rr := newResponseRecorder(b)

		m.HandleRequest(rr, req)

		if want := 200; rr.status != want {
			t.Errorf("status code got: %v want: %v", rr.status, want)
		}

		wantHeaders := map[string][]string{}
		if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
			t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
		}

		if got, want := b.String(), ""; got != want {
			t.Errorf("response body got: %q want: %q", got, want)
		}
	})

	t.Run("Desired Behavior", func(t *testing.T) {
		t.Skip()
		b := &strings.Builder{}
		rr := newResponseRecorder(b)

		m.HandleRequest(rr, req)

		if want := 500; rr.status != want {
			t.Errorf("status code got: %v want: %v", rr.status, want)
		}

		wantHeaders := map[string][]string{}
		if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
			t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
		}

		if got, want := b.String(), "Internal Server Error"; got != want {
			t.Errorf("response body got: %q want: %q", got, want)
		}
	})
}
