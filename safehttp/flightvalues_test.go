// Copyright 2020 Google LLC
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

package safehttp_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

type safeHeadersInterceptor struct{}

func (ip *safeHeadersInterceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	// We claim the header here in order to protect it from being tampered. The
	// only way to set it is through a helper method exposed by this package. It
	// only allows for setting safe values.
	setter := w.Header().Claim("Super-Safe-Header")
	safehttp.FlightValues(r.Context()).Put(safeHeaderKey{}, setter)
	return safehttp.NotWritten()
}

func (ip *safeHeadersInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

func (ip *safeHeadersInterceptor) Match(_ safehttp.InterceptorConfig) bool {
	// This interceptor does not offer any configuration options.
	return false
}

type safeHeaderKey struct{}

func SetHeaderSafely(ctx context.Context, level int) {
	var value string
	switch level {
	case 0:
		value = "Safe"
	case 1:
		value = "VerySafe"
	case 2:
		value = "VeryVerySafe"
	default:
		value = "Safe"
	}
	setter := safehttp.FlightValues(ctx).Get(safeHeaderKey{}).(func([]string))
	setter([]string{value})
}

func handlerInteractingWithTheInterceptor(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	f, err := req.URL().Query()
	if err != nil {
		panic(err)
	}
	safety := f.Int64("level", 0)
	SetHeaderSafely(req.Context(), int(safety))

	return w.Write(safehtml.HTMLEscaped(fmt.Sprintf("Safety header set to %v", safety)))
}

func TestHandlerInteractingWithInterceptor(t *testing.T) {
	mb := safehttp.NewServeMuxConfig(nil)
	mb.Intercept(&safeHeadersInterceptor{})
	m := mb.Mux()

	m.Handle("/safety", safehttp.MethodGet, safehttp.HandlerFunc(handlerInteractingWithTheInterceptor))

	rr := httptest.NewRecorder()

	req := httptest.NewRequest(safehttp.MethodGet, "https://foo.com/safety?level=2", nil)
	m.ServeHTTP(rr, req)

	if got, want := rr.Code, safehttp.StatusOK; got != int(want) {
		t.Errorf("rr.Code got: %v want: %v", got, want)
	}

	want := `Safety header set to 2`
	if diff := cmp.Diff(want, rr.Body.String()); diff != "" {
		t.Errorf("response body diff (-want,+got): \n%s\ngot %q, want %q", diff, rr.Body.String(), want)
	}

	wantHeaders := map[string][]string{
		"Content-Type":      {"text/html; charset=utf-8"},
		"Super-Safe-Header": {"VeryVerySafe"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header mismatch (-want +got):\n%s", diff)
	}
}
