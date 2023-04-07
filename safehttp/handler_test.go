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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

var BarHandler = safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	if !strings.HasPrefix(r.URL().Path(), "/bar") {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	return w.Write(safehtml.HTMLEscaped("Hello!"))
})

func TestStripPrefix(t *testing.T) {
	mux := safehttp.NewServeMuxConfig(nil).Mux()

	mux.Handle("/bar", safehttp.MethodGet, BarHandler)
	mux.Handle("/more/bar", safehttp.MethodGet, safehttp.StripPrefix("/more", BarHandler))

	r := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, r)

	rStrip := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/more/bar", nil)
	rwStrip := httptest.NewRecorder()
	mux.ServeHTTP(rwStrip, rStrip)

	if rwStrip.Code != rw.Code {
		t.Errorf("Code got %v, want %v", rwStrip.Code, rw.Code)
	}

	if diff := cmp.Diff(rw.Header(), rwStrip.Header()); diff != "" {
		t.Errorf("Header() mismatch (-want +got):\n%s", diff)
	}

	if got := rwStrip.Body.String(); got != rw.Body.String() {
		t.Errorf("response body: got %q want %q", got, rw.Body.String())
	}
}

func TestStripPrefixPanic(t *testing.T) {
	mux := safehttp.NewServeMuxConfig(nil).Mux()

	mux.Handle("/bar", safehttp.MethodGet, BarHandler)
	mux.Handle("/more/bar", safehttp.MethodGet, safehttp.StripPrefix("/badprefix", BarHandler))

	r := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/more/bar", nil)
	rw := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf("expected panic")
	}()
	mux.ServeHTTP(rw, r)
}
