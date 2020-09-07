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

package staticheaders_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/staticheaders"
	"github.com/google/safehtml"
)

type dispatcher struct{}

func (dispatcher) ContentType(resp safehttp.Response) (string, error) {
	switch resp.(type) {
	case safehtml.HTML, *template.Template:
		return "text/html; charset=utf-8", nil
	case safehttp.JSONResponse:
		return "application/json; charset=utf-8", nil
	default:
		return "", errors.New("not a safe response")
	}
}

func (dispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (dispatcher) WriteJSON(rw http.ResponseWriter, resp safehttp.JSONResponse) error {
	obj, err := json.Marshal(resp.Data)
	if err != nil {
		panic("invalid json")
	}
	_, err = rw.Write(obj)
	return err
}

func (dispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
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
	status safehttp.StatusCode
}

func newResponseRecorder(w io.Writer) *responseRecorder {
	return &responseRecorder{header: http.Header{}, writer: w, status: http.StatusOK}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = safehttp.StatusCode(statusCode)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

func TestServeMuxInstallStaticHeaders(t *testing.T) {
	mux := safehttp.NewServeMux(dispatcher{}, "foo.com")

	mux.Install(staticheaders.Plugin{})
	handler := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
	})
	mux.Handle("/asdf", safehttp.MethodGet, handler)

	b := strings.Builder{}
	rr := newResponseRecorder(&b)

	req := httptest.NewRequest(http.MethodGet, "https://foo.com/asdf", nil)

	mux.ServeHTTP(rr, req)

	if want := safehttp.StatusOK; rr.status != want {
		t.Errorf("rr.status got: %v want: %v", rr.status, want)
	}

	wantHeaders := map[string][]string{
		"Content-Type":           {"text/html; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
		"X-Xss-Protection":       {"0"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.header)); diff != "" {
		t.Errorf("rr.header mismatch (-want +got):\n%s", diff)
	}

	if got, want := b.String(), "&lt;h1&gt;Hello World!&lt;/h1&gt;"; got != want {
		t.Errorf("b.String() got: %v want: %v", got, want)
	}
}
