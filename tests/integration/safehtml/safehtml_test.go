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

package safehtml_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

type dispatcher struct{}

func (dispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (dispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

func TestHandleRequestWrite(t *testing.T) {
	m := safehttp.NewMachinery(func(rw safehttp.ResponseWriter, _ *safehttp.IncomingRequest) safehttp.Result {
		return rw.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
	}, &dispatcher{})

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	m.HandleRequest(recorder, req)

	resp := recorder.Result()

	want := "&lt;h1&gt;Escaped, so not really a heading&lt;/h1&gt;"
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(resp.Body) got err: %v, want nil", err)
	}

	if got := string(body); got != want {
		t.Errorf("resp.Body: got = %q, want %q", got, want)
	}
}

func TestHandleRequestWriteTemplate(t *testing.T) {
	m := safehttp.NewMachinery(func(rw safehttp.ResponseWriter, _ *safehttp.IncomingRequest) safehttp.Result {
		return rw.WriteTemplate(template.Must(template.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
	}, &dispatcher{})

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	m.HandleRequest(recorder, req)

	resp := recorder.Result()

	want := "<h1>This is an actual heading, though.</h1>"
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(resp.Body) got err: %v, want nil", err)
	}

	if got := string(body); got != want {
		t.Errorf("resp.Body: got = %q, want %q", got, want)
	}
}
