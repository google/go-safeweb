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

package errors_test

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/safehtml"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml/template"
)

var myErrorTmpl = template.Must(template.New("not found").Parse(`<h1>Error: {{ .Code }}</h1>
<p>{{ .Message }}</p>
`))

type myError struct {
	safehttp.StatusCode
	Message string
}

type myDispatcher struct {
	safehttp.DefaultDispatcher
}

func (m myDispatcher) Error(rw http.ResponseWriter, resp safehttp.ErrorResponse) error {
	if x, ok := resp.(myError); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(int(x.StatusCode))
		return myErrorTmpl.Execute(rw, x)
	}
	return m.DefaultDispatcher.Error(rw, resp)
}

// TestCustomErrors tests a scenario where custom errors are implemented using a
// custom dispatcher implementation. XSRF interceptor is added to check that
// error responses go through the commit phase.
func TestCustomErrors(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(myDispatcher{})
	mux := mb.Mux()

	mux.Handle("/compute", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		qs, err := r.URL().Query()
		if err != nil {
			return w.WriteError(safehttp.StatusBadRequest)
		}
		a := qs.Int64("a", math.MaxInt64)
		if qs.Err() != nil || a == math.MaxInt64 {
			return w.WriteError(myError{StatusCode: safehttp.StatusBadRequest, Message: "missing parameter 'a'"})
		}
		if a > 10 {
			return w.WriteError(myError{StatusCode: safehttp.StatusNotImplemented, Message: "we can't process queries with large numbers yet"})
		}
		return w.Write(safehtml.HTMLEscaped(fmt.Sprintf("Result: %d", a*a)))
	}))

	t.Run("correct request", func(t *testing.T) {
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/compute?a=3", nil)
		mux.ServeHTTP(rr, req)

		if got, want := rr.Code, safehttp.StatusOK; got != int(want) {
			t.Errorf("rr.Code got: %v want: %v", got, want)
		}
		want := "Result: 9"
		if diff := cmp.Diff(want, rr.Body.String()); diff != "" {
			t.Errorf("response body diff (-want,+got): \n%s\ngot %q, want %q", diff, rr.Body.String(), want)
		}
	})

	t.Run("missing parameter", func(t *testing.T) {
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/compute?foo=3", nil)
		mux.ServeHTTP(rr, req)

		if got, want := rr.Code, safehttp.StatusBadRequest; got != int(want) {
			t.Errorf("rr.Code got: %v want: %v", got, want)
		}

		want := `<h1>Error: Bad Request</h1>
<p>missing parameter &#39;a&#39;</p>
`
		if diff := cmp.Diff(want, rr.Body.String()); diff != "" {
			t.Errorf("response body diff (-want,+got): \n%s\ngot %q, want %q", diff, rr.Body.String(), want)
		}
	})

}

type interceptor struct {
	errBefore bool
}

func (it interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if it.errBefore {
		return w.WriteError(myError{StatusCode: safehttp.StatusForbidden, Message: "forbidden in Before"})
	}
	return safehttp.NotWritten()
}

func (interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}

func (interceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}

func TestCustomErrorsInBefore(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(myDispatcher{})
	mb.Intercept(interceptor{errBefore: true})
	mux := mb.Mux()

	mux.Handle("/compute", safehttp.MethodGet, safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("the handler code doesn't matter in this test case"))
	}))

	t.Run("error in Before", func(t *testing.T) {
		rr := httptest.NewRecorder()

		req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/compute?a=3", nil)
		mux.ServeHTTP(rr, req)

		if got, want := rr.Code, safehttp.StatusForbidden; got != int(want) {
			t.Errorf("rr.Code got: %v want: %v", got, want)
		}
		want := `<h1>Error: Forbidden</h1>
<p>forbidden in Before</p>
`
		if diff := cmp.Diff(want, rr.Body.String()); diff != "" {
			t.Errorf("response body diff (-want,+got): \n%s\ngot %q, want %q", diff, rr.Body.String(), want)
		}
	})
}
