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
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp/safehttptest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

func TestMuxOneHandlerOneRequest(t *testing.T) {
	var test = []struct {
		name       string
		req        *http.Request
		wantStatus safehttp.StatusCode
		wantHeader map[string][]string
		wantBody   string
	}{
		{
			name:       "Valid Request",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://foo.com/", nil),
			wantStatus: safehttp.StatusOK,
			wantHeader: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name:       "Invalid Method",
			req:        httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", nil),
			wantStatus: safehttp.StatusMethodNotAllowed,
			wantHeader: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Method Not Allowed\n",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mb := &safehttp.ServeMuxConfig{}

			h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mb.Handle("/", safehttp.MethodGet, h)

			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)

			mux := mb.Mux()
			mux.ServeHTTP(rw, tt.req)

			if rw.Status() != tt.wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeader, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := b.String(); got != tt.wantBody {
				t.Errorf("response body: got %q want %q", got, tt.wantBody)
			}
		})
	}
}

func TestMuxServeTwoHandlers(t *testing.T) {
	var tests = []struct {
		name        string
		req         *http.Request
		hf          safehttp.Handler
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "GET Handler",
			req:  httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil),
			hf: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World! GET</h1>"))
			}),
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World! GET&lt;/h1&gt;",
		},
		{
			name: "POST Handler",
			req:  httptest.NewRequest(safehttp.MethodPost, "http://foo.com/bar", nil),
			hf: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World! POST</h1>"))
			}),
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World! POST&lt;/h1&gt;",
		},
	}

	mb := &safehttp.ServeMuxConfig{}
	mb.Handle("/bar", safehttp.MethodGet, tests[0].hf)
	mb.Handle("/bar", safehttp.MethodPost, tests[1].hf)
	mux := mb.Mux()

	for _, test := range tests {
		b := &strings.Builder{}
		rw := safehttptest.NewTestResponseWriter(b)
		mux.ServeHTTP(rw, test.req)
		if want := test.wantStatus; rw.Status() != want {
			t.Errorf("rw.Status(): got %v want %v", rw.Status(), want)
		}

		if diff := cmp.Diff(test.wantHeaders, map[string][]string(rw.Header())); diff != "" {
			t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
		}

		if got, want := b.String(), test.wantBody; got != want {
			t.Errorf("response body: got %q want %q", got, want)
		}
	}
}

func TestMuxRegisterCorrectHandlerAllPaths(t *testing.T) {
	var tests = []struct {
		name     string
		req      *http.Request
		hf       safehttp.Handler
		wantBody string
	}{
		{
			name: "GET Handler",
			req:  httptest.NewRequest(safehttp.MethodGet, "http://foo.com/get", nil),
			hf: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("GET handler for /get"))
			}),
			wantBody: "GET handler for /get",
		},
		{
			name: "GET Handler #2",
			req:  httptest.NewRequest(safehttp.MethodGet, "http://foo.com/get2", nil),
			hf: safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("GET handler for /get2"))
			}),
			wantBody: "GET handler for /get2",
		},
	}

	mb := &safehttp.ServeMuxConfig{}
	mb.Handle("/get", safehttp.MethodGet, tests[0].hf)
	mb.Handle("/get2", safehttp.MethodGet, tests[1].hf)
	mux := mb.Mux()

	for _, test := range tests {
		b := &strings.Builder{}
		rw := safehttptest.NewTestResponseWriter(b)
		mux.ServeHTTP(rw, test.req)

		if got, want := b.String(), test.wantBody; got != want {
			t.Errorf("response body: got %q want %q", got, want)
		}
	}
}

func TestMuxHandleSameMethodTwice(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}

	registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mb.Handle("/bar", safehttp.MethodGet, registeredHandler)

	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`mux.Handle("/bar", MethodGet, registeredHandler) expected panic`)
	}()

	mb.Handle("/bar", safehttp.MethodGet, registeredHandler)
	mb.Mux()
}

type setHeaderInterceptor struct {
	name  string
	value string
}

func (p setHeaderInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	w.Header().Set(p.name, p.value)
	return safehttp.NotWritten()
}

func (p setHeaderInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

type internalErrorInterceptor struct{}

func (internalErrorInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	return w.WriteError(safehttp.StatusInternalServerError)
}

func (internalErrorInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

type claimHeaderInterceptor struct {
	headerToClaim string
}

type claimCtxKey struct{}

func (p *claimHeaderInterceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	f := w.Header().Claim(p.headerToClaim)
	r.SetContext(context.WithValue(r.Context(), claimCtxKey{}, f))
	return safehttp.NotWritten()
}

func (p *claimHeaderInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

func claimInterceptorSetHeader(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, value string) {
	f := r.Context().Value(claimCtxKey{}).(func([]string))
	f([]string{value})
}

type committerInterceptor struct{}

func (committerInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	return safehttp.NotWritten()
}

func (committerInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	w.Header().Set("foo", "bar")
}

type setHeaderErroringInterceptor struct{}

func (setHeaderErroringInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	return w.WriteError(safehttp.StatusForbidden)
}

func (setHeaderErroringInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	w.Header().Set("name", "foo")
}

func TestMuxInterceptors(t *testing.T) {
	tests := []struct {
		name        string
		mux         *safehttp.ServeMux
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name: "Install ServeMux Interceptor before handler registration",
			mux: func() *safehttp.ServeMux {
				mb := &safehttp.ServeMuxConfig{}
				mb.Intercept(setHeaderInterceptor{
					name:  "Foo",
					value: "bar",
				})

				registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
				})
				mb.Handle("/bar", safehttp.MethodGet, registeredHandler)
				return mb.Mux()
			}(),
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Foo":          {"bar"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name: "Install Interrupting Interceptor",
			mux: func() *safehttp.ServeMux {
				mb := &safehttp.ServeMuxConfig{}
				mb.Intercept(internalErrorInterceptor{})

				registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
				})
				mb.Handle("/bar", safehttp.MethodGet, registeredHandler)

				return mb.Mux()
			}(),
			wantStatus: safehttp.StatusInternalServerError,
			wantHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
			wantBody: "Internal Server Error\n",
		},
		{
			name: "Handler Communication With ServeMux Interceptor",
			mux: func() *safehttp.ServeMux {
				mb := &safehttp.ServeMuxConfig{}
				mb.Intercept(&claimHeaderInterceptor{headerToClaim: "Foo"})

				registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					claimInterceptorSetHeader(w, r, "bar")
					return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
				})
				mb.Handle("/bar", safehttp.MethodGet, registeredHandler)

				return mb.Mux()
			}(),
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Foo":          {"bar"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name: "Commit phase sets header",
			mux: func() *safehttp.ServeMux {
				mb := &safehttp.ServeMuxConfig{}
				mb.Intercept(committerInterceptor{})

				registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
					return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
				})
				mb.Handle("/bar", safehttp.MethodGet, registeredHandler)

				return mb.Mux()
			}(),
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Foo":          {"bar"},
				"Content-Type": {"text/html; charset=utf-8"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)

			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil)

			tt.mux.ServeHTTP(rw, req)

			if rw.Status() != tt.wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := b.String(); got != tt.wantBody {
				t.Errorf("response body: got %q want %q", got, tt.wantBody)
			}
		})
	}
}

type setHeaderConfig struct {
	name  string
	value string
}

func (setHeaderConfig) Match(i safehttp.Interceptor) bool {
	_, ok := i.(setHeaderConfigInterceptor)
	return ok
}

type setHeaderConfigInterceptor struct{}

func (p setHeaderConfigInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	name := "Pizza"
	value := "Hawaii"
	if c, ok := cfg.(setHeaderConfig); ok {
		name = c.name
		value = c.value
	}
	w.Header().Set(name, value)
	return safehttp.NotWritten()
}

func (p setHeaderConfigInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	name := "Commit-Pizza"
	value := "Hawaii"
	if c, ok := cfg.(setHeaderConfig); ok {
		name = "Commit-" + c.name
		value = c.value
	}
	w.Header().Set(name, value)
}

type noInterceptorConfig struct{}

func (noInterceptorConfig) Match(i safehttp.Interceptor) bool {
	return false
}

func TestMuxInterceptorConfigs(t *testing.T) {
	tests := []struct {
		name        string
		config      safehttp.InterceptorConfig
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:       "SetHeaderInterceptor with config",
			config:     setHeaderConfig{name: "Foo", value: "Bar"},
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Commit-Foo":   {"Bar"},
				"Foo":          {"Bar"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name:       "SetHeaderInterceptor with mismatching config",
			config:     noInterceptorConfig{},
			wantStatus: safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Pizza":        {"Hawaii"},
				"Commit-Pizza": {"Hawaii"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &safehttp.ServeMuxConfig{}
			mb.Intercept(setHeaderConfigInterceptor{})

			registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mb.Handle("/bar", safehttp.MethodGet, registeredHandler, tt.config)

			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)

			req := httptest.NewRequest("GET", "http://foo.com/bar", nil)

			mux := mb.Mux()
			mux.ServeHTTP(rw, req)

			if rw.Status() != tt.wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := b.String(); got != tt.wantBody {
				t.Errorf("response body: got %q want %q", got, tt.wantBody)
			}
		})
	}
}

type interceptorOne struct{}

func (interceptorOne) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	w.Header().Set("pizza", "diavola")
	return safehttp.NotWritten()
}

func (interceptorOne) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	if w.Header().Get("Commit2") != "b" {
		panic("server bug")
	}
	w.Header().Set("Commit1", "a")
}

type interceptorTwo struct{}

func (interceptorTwo) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	if w.Header().Get("pizza") != "diavola" {
		panic("server bug")
	}
	w.Header().Set("spaghetti", "bolognese")
	return safehttp.NotWritten()
}

func (interceptorTwo) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	if w.Header().Get("Commit3") != "c" {
		panic("server bug")
	}
	w.Header().Set("Commit2", "b")
}

type interceptorThree struct{}

func (interceptorThree) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	if w.Header().Get("spaghetti") != "bolognese" {
		panic("server bug")
	}
	w.Header().Set("dessert", "tiramisu")
	return safehttp.NotWritten()
}

func (interceptorThree) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	if w.Header().Get("Dessert") != "tiramisu" {
		panic("server bug")
	}
	w.Header().Set("Commit3", "c")
}

func TestMuxDeterministicInterceptorOrder(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}
	mb.Intercept(interceptorOne{})
	mb.Intercept(interceptorTwo{})
	mb.Intercept(interceptorThree{})

	registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mb.Handle("/bar", safehttp.MethodGet, registeredHandler)

	b := &strings.Builder{}
	rw := safehttptest.NewTestResponseWriter(b)

	req := httptest.NewRequest("GET", "http://foo.com/bar", nil)

	mux := mb.Mux()
	mux.ServeHTTP(rw, req)

	if want := safehttp.StatusOK; rw.Status() != want {
		t.Errorf("rw.Status(): got %v want %v", rw.Status(), want)
	}
	wantHeaders := map[string][]string{
		"Dessert":      {"tiramisu"},
		"Pizza":        {"diavola"},
		"Spaghetti":    {"bolognese"},
		"Commit1":      {"a"},
		"Commit2":      {"b"},
		"Commit3":      {"c"},
		"Content-Type": {"text/html; charset=utf-8"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}
	if got, want := b.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf(`response body: got %q want %q`, got, want)
	}
}

func TestMuxHandlerReturnsNotWritten(t *testing.T) {
	mb := &safehttp.ServeMuxConfig{}
	h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return safehttp.NotWritten()
	})
	mb.Handle("/bar", safehttp.MethodGet, h)
	req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil)

	b := &strings.Builder{}
	rw := safehttptest.NewTestResponseWriter(b)

	mux := mb.Mux()
	mux.ServeHTTP(rw, req)

	if want := safehttp.StatusNoContent; rw.Status() != want {
		t.Errorf("rw.Status(): got %v want %v", rw.Status(), want)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}
	if got := b.String(); got != "" {
		t.Errorf(`response body got: %q want: ""`, got)
	}
}
