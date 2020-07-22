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

package safehttp

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	status200OK     = "200 OK"
	status400BadReq = "400 Bad Request"
)

type dispatcher struct{}

func (d *dispatcher) Write(rw http.ResponseWriter, resp Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (d *dispatcher) ExecuteTemplate(rw http.ResponseWriter, t Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

// A helper function that returns a parsed Form or an empty Form and any errors
// that occured during parsing
func getParsedForm(r *IncomingRequest) (*Form, error) {
	if r.req.Method == "GET" {
		f, err := r.QueryForm()
		return f, err
	}

	if !strings.HasPrefix(r.req.Header.Get("Content-Type"), "multipart/form-data") {
		f, err := r.PostForm()
		return f, err
	}
	mf, err := r.MultipartForm(32 << 20)
	return mf.Form, err
}

func TestValidInt(t *testing.T) {
	multipartReqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"pizza\"\r\n" +
		"\r\n" +
		"10\r\n" +
		"--123--\r\n"
	getReq := httptest.NewRequest("GET", "/?pizza=10", nil)
	postReq := httptest.NewRequest("POST", "/", strings.NewReader("pizza=10"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	multipartReq := httptest.NewRequest("POST", "/", strings.NewReader(multipartReqBody))
	multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
	tests := []struct {
		name    string
		req     *http.Request
		formVal int
	}{
		{
			name:    "valid int in GET request",
			req:     getReq,
			formVal: 10,
		},
		{
			name:    "valid int in POST non-multipart request",
			req:     postReq,
			formVal: 10,
		},
		{
			name:    "valid int in POST multipart request",
			req:     multipartReq,
			formVal: 10,
		},
	}

	for _, test := range tests {
		m := NewMachinery(func(rw ResponseWriter, ir *IncomingRequest) Result {
			form, err := getParsedForm(ir)
			if err != nil {
				t.Errorf(`getParsedForm: got "%v", want nil`, err)
			}
			want := test.formVal
			got := form.Int("pizza", 0)
			if err := form.Error(); err != nil {
				t.Errorf(`form.Error: got "%v", want nil`, err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("form.Int: got %v, want %v, diff (-want +got): \n%s", got, want, diff)
			}
			return Result{}
		}, &dispatcher{})
		recorder := httptest.NewRecorder()
		m.HandleRequest(recorder, test.req)
		if respStatus := recorder.Result().Status; respStatus != status200OK {
			t.Errorf("response status: got %s, want %s", respStatus, status200OK)
		}
	}
}

func TestValidIntSlice(t *testing.T) {
	multipartReqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"pizza\"\r\n" +
		"\r\n" +
		"10\r\n" +
		"--123\r\n" +
		"Content-Disposition: form-data; name=\"pizza\"\r\n" +
		"\r\n" +
		"4\r\n" +
		"--123--\r\n"
	getReq := httptest.NewRequest("GET", "/?pizza=10&pizza=4", nil)
	postReq := httptest.NewRequest("POST", "/", strings.NewReader("pizza=10&pizza=4"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	multipartReq := httptest.NewRequest("POST", "/", strings.NewReader(multipartReqBody))
	multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
	tests := []struct {
		name    string
		req     *http.Request
		formVal []int
	}{
		{
			name:    "valid int slice in GET request",
			req:     getReq,
			formVal: []int{10, 4},
		},
		{
			name:    "valid int slice in POST non-multipart request",
			req:     postReq,
			formVal: []int{10, 4},
		},
		{
			name:    "valid int slice in POST multipart request",
			req:     multipartReq,
			formVal: []int{10, 4},
		},
	}
	for _, test := range tests {
		m := NewMachinery(func(rw ResponseWriter, ir *IncomingRequest) Result {
			form, err := getParsedForm(ir)
			if err != nil {
				t.Errorf(`getParsedForm: got "%v", want nil`, err)
			}
			want := test.formVal
			var got []int
			form.Slice(&got, "pizza")
			if err := form.Error(); err != nil {
				t.Errorf(`form.Error: got "%v", want nil`, err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("form.Slice: got %v, want %v, diff (-want +got): \n%s", got, want, diff)
			}
			return Result{}
		}, &dispatcher{})
		recorder := httptest.NewRecorder()
		m.HandleRequest(recorder, test.req)
		if respStatus := recorder.Result().Status; respStatus != status200OK {
			t.Errorf("response status: got %s, want %s", respStatus, status200OK)
		}
	}
}

// TODO(@mihalimara22): Add more tests
