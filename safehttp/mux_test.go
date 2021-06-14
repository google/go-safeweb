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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
			mb := safehttp.NewServeMuxConfig(nil)

			h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mb.Handle("/", safehttp.MethodGet, h)

			rw := httptest.NewRecorder()

			mux := mb.Mux()
			mux.ServeHTTP(rw, tt.req)

			if rw.Code != int(tt.wantStatus) {
				t.Errorf("rw.Code: got %v want %v", rw.Code, tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeader, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := rw.Body.String(); got != tt.wantBody {
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

	mb := safehttp.NewServeMuxConfig(nil)
	mb.Handle("/bar", safehttp.MethodGet, tests[0].hf)
	mb.Handle("/bar", safehttp.MethodPost, tests[1].hf)
	mux := mb.Mux()

	for _, test := range tests {
		rw := httptest.NewRecorder()
		mux.ServeHTTP(rw, test.req)
		if want := int(test.wantStatus); rw.Code != want {
			t.Errorf("rw.Code: got %v want %v", rw.Code, want)
		}

		if diff := cmp.Diff(test.wantHeaders, map[string][]string(rw.Header())); diff != "" {
			t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
		}

		if got, want := rw.Body.String(), test.wantBody; got != want {
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

	mb := safehttp.NewServeMuxConfig(nil)
	mb.Handle("/get", safehttp.MethodGet, tests[0].hf)
	mb.Handle("/get2", safehttp.MethodGet, tests[1].hf)
	mux := mb.Mux()

	for _, test := range tests {
		rw := httptest.NewRecorder()
		mux.ServeHTTP(rw, test.req)

		if got, want := rw.Body.String(), test.wantBody; got != want {
			t.Errorf("response body: got %q want %q", got, want)
		}
	}
}

func TestMuxHandleSameMethodTwice(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)

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

func (setHeaderInterceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}

type internalErrorInterceptor struct{}

func (internalErrorInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	return w.WriteError(safehttp.StatusInternalServerError)
}

func (internalErrorInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

func (internalErrorInterceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}

type claimHeaderInterceptor struct {
	headerToClaim string
}

type claimCtxKey struct{}

func (p *claimHeaderInterceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	f := w.Header().Claim(p.headerToClaim)
	safehttp.FlightValues(r.Context()).Put(claimCtxKey{}, f)
	return safehttp.NotWritten()
}

func (p *claimHeaderInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

func (claimHeaderInterceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}

func claimInterceptorSetHeader(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, value string) {
	f := safehttp.FlightValues(r.Context()).Get(claimCtxKey{}).(func([]string))
	f([]string{value})
}

type committerInterceptor struct{}

func (committerInterceptor) Before(w safehttp.ResponseWriter, _ *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	return safehttp.NotWritten()
}

func (committerInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	w.Header().Set("foo", "bar")
}

func (committerInterceptor) Match(safehttp.InterceptorConfig) bool {
	return false
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
				mb := safehttp.NewServeMuxConfig(nil)
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
				mb := safehttp.NewServeMuxConfig(nil)
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
				mb := safehttp.NewServeMuxConfig(nil)
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
				mb := safehttp.NewServeMuxConfig(nil)
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
			rw := httptest.NewRecorder()

			req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil)

			tt.mux.ServeHTTP(rw, req)

			if rw.Code != int(tt.wantStatus) {
				t.Errorf("rw.Code: got %v want %v", rw.Code, tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := rw.Body.String(); got != tt.wantBody {
				t.Errorf("response body: got %q want %q", got, tt.wantBody)
			}
		})
	}
}

type setHeaderConfig struct {
	name  string
	value string
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

func (setHeaderConfigInterceptor) Match(cfg safehttp.InterceptorConfig) bool {
	_, ok := cfg.(setHeaderConfig)
	return ok
}

type noInterceptorConfig struct{}

type wrappedInterceptor struct {
	w safehttp.Interceptor
}

func (wi wrappedInterceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	return wi.w.Before(w, r, cfg)
}

func (wi wrappedInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	wi.w.Commit(w, r, resp, cfg)
}

func (wi wrappedInterceptor) Match(cfg safehttp.InterceptorConfig) bool {
	return wi.w.Match(cfg)
}

func (noInterceptorConfig) Match(i safehttp.Interceptor) bool {
	return false
}

func TestMuxInterceptorConfigs(t *testing.T) {
	tests := []struct {
		name        string
		interceptor safehttp.Interceptor
		config      safehttp.InterceptorConfig
		wantStatus  safehttp.StatusCode
		wantHeaders map[string][]string
		wantBody    string
	}{
		{
			name:        "SetHeaderInterceptor with config",
			interceptor: setHeaderConfigInterceptor{},
			config:      setHeaderConfig{name: "Foo", value: "Bar"},
			wantStatus:  safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Commit-Foo":   {"Bar"},
				"Foo":          {"Bar"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name:        "Wrapped SetHeaderInterceptor with config",
			interceptor: wrappedInterceptor{w: setHeaderConfigInterceptor{}},
			config:      setHeaderConfig{name: "Foo", value: "Bar"},
			wantStatus:  safehttp.StatusOK,
			wantHeaders: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
				"Commit-Foo":   {"Bar"},
				"Foo":          {"Bar"},
			},
			wantBody: "&lt;h1&gt;Hello World!&lt;/h1&gt;",
		},
		{
			name:        "SetHeaderInterceptor with mismatching config",
			interceptor: setHeaderConfigInterceptor{},
			config:      noInterceptorConfig{},
			wantStatus:  safehttp.StatusOK,
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
			mb := safehttp.NewServeMuxConfig(nil)
			mb.Intercept(tt.interceptor)

			registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mb.Handle("/bar", safehttp.MethodGet, registeredHandler, tt.config)

			rw := httptest.NewRecorder()

			req := httptest.NewRequest("GET", "http://foo.com/bar", nil)

			mux := mb.Mux()
			mux.ServeHTTP(rw, req)

			if rw.Code != int(tt.wantStatus) {
				t.Errorf("rw.Code: got %v want %v", rw.Code, tt.wantStatus)
			}

			if diff := cmp.Diff(tt.wantHeaders, map[string][]string(rw.Header())); diff != "" {
				t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
			}

			if got := rw.Body.String(); got != tt.wantBody {
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

func (interceptorOne) Match(safehttp.InterceptorConfig) bool {
	return false
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

func (interceptorTwo) Match(safehttp.InterceptorConfig) bool {
	return false
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

func (interceptorThree) Match(safehttp.InterceptorConfig) bool {
	return false
}

func TestMuxDeterministicInterceptorOrder(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)
	mb.Intercept(interceptorOne{})
	mb.Intercept(interceptorTwo{})
	mb.Intercept(interceptorThree{})

	registeredHandler := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mb.Handle("/bar", safehttp.MethodGet, registeredHandler)

	rw := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "http://foo.com/bar", nil)

	mux := mb.Mux()
	mux.ServeHTTP(rw, req)

	if want := safehttp.StatusOK; rw.Code != int(want) {
		t.Errorf("rw.Code: got %v want %v", rw.Code, want)
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
	if got, want := rw.Body.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf(`response body: got %q want %q`, got, want)
	}
}

func TestMuxHandlerReturnsNotWritten(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)
	h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return safehttp.NotWritten()
	})
	mb.Handle("/bar", safehttp.MethodGet, h)
	req := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil)

	rw := httptest.NewRecorder()

	mux := mb.Mux()
	mux.ServeHTTP(rw, req)

	if want := safehttp.StatusNoContent; rw.Code != int(want) {
		t.Errorf("rw.Code: got %v want %v", rw.Code, want)
	}
	if diff := cmp.Diff(map[string][]string{}, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}
	if got := rw.Body.String(); got != "" {
		t.Errorf(`response body got: %q want: ""`, got)
	}
}

func TestMuxMethodNotAllowedDefaults(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)

	h := safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		panic("not tested")
	})
	mb.Handle("/", safehttp.MethodGet, h)

	rw := httptest.NewRecorder()

	mux := mb.Mux()
	mux.ServeHTTP(rw, httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", nil))

	if got, want := rw.Code, int(safehttp.StatusMethodNotAllowed); got != want {
		t.Errorf("rw.Code: got %v want %v", got, want)
	}

	wantHeader := map[string][]string{
		"Content-Type":           {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	if diff := cmp.Diff(wantHeader, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}

	wantBody := "Method Not Allowed\n"
	if got := rw.Body.String(); got != wantBody {
		t.Errorf("response body: got %q want %q", got, wantBody)
	}
}

type methodNotAllowedError struct {
	message string
}

func (err *methodNotAllowedError) Code() safehttp.StatusCode {
	return safehttp.StatusMethodNotAllowed
}

type methodNotAllowedDispatcher struct {
	safehttp.DefaultDispatcher
}

func (d *methodNotAllowedDispatcher) Error(rw http.ResponseWriter, resp safehttp.ErrorResponse) error {
	x := resp.(*methodNotAllowedError)
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(int(resp.Code()))
	_, err := io.WriteString(rw, "<h1>"+http.StatusText(int(resp.Code()))+"</h1>"+"<p>"+x.message+"</p>")
	return err
}

type methodNotAllowedInterceptor struct{}

func (ip *methodNotAllowedInterceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, ipcfg safehttp.InterceptorConfig) safehttp.Result {
	cfg := ipcfg.(methodNotAllowedInterceptorConfig)
	w.Header().Set("Before-Interceptor", cfg.before)
	return safehttp.NotWritten()
}

// Commit runs before the response is written by the Dispatcher. If an error
// is written to the ResponseWriter, then the Commit phases from the
// remaining interceptors won't execute.
func (ip *methodNotAllowedInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, ipcfg safehttp.InterceptorConfig) {
	cfg := ipcfg.(methodNotAllowedInterceptorConfig)
	w.Header().Set("Commit-Interceptor", cfg.commit)
}

func (*methodNotAllowedInterceptor) Match(cfg safehttp.InterceptorConfig) bool {
	_, ok := cfg.(methodNotAllowedInterceptorConfig)
	return ok
}

type methodNotAllowedInterceptorConfig struct {
	before, commit string
}

func TestMuxMethodNotAllowedCustom(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(&methodNotAllowedDispatcher{})
	mb.Intercept(&methodNotAllowedInterceptor{})

	mb.Handle("/", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		panic("not tested")
	}))
	mb.HandleMethodNotAllowed(safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
		return rw.WriteError(&methodNotAllowedError{"custom message"})
	}), methodNotAllowedInterceptorConfig{before: "foo", commit: "bar"})

	rw := httptest.NewRecorder()

	mux := mb.Mux()
	mux.ServeHTTP(rw, httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", nil))

	if got, want := rw.Code, int(safehttp.StatusMethodNotAllowed); got != want {
		t.Errorf("rw.Code: got %v want %v", got, want)
	}

	wantHeader := map[string][]string{
		"Content-Type":       {"text/html; charset=utf-8"},
		"Before-Interceptor": {"foo"},
		"Commit-Interceptor": {"bar"},
	}
	if diff := cmp.Diff(wantHeader, map[string][]string(rw.Header())); diff != "" {
		t.Errorf("rw.Header() mismatch (-want +got):\n%s", diff)
	}

	wantBody := "<h1>Method Not Allowed</h1><p>custom message</p>"
	if got := rw.Body.String(); got != wantBody {
		t.Errorf("response body: got %q want %q", got, wantBody)
	}
}
