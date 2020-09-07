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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"text/template"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

// This file contains a test dispatcher and a responserecorder
// that implements a http.ResponseWriter. These are used for
// testing.

type testDispatcher struct{}

func (testDispatcher) ContentType(resp safehttp.Response) (string, error) {
	switch resp.(type) {
	case safehtml.HTML, *template.Template:
		return "text/html; charset=utf-8", nil
	case safehttp.JSONResponse:
		return "application/json; charset=utf-8", nil
	default:
		return "", errors.New("not a safe response")
	}
}

func (testDispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (testDispatcher) WriteJSON(rw http.ResponseWriter, resp safehttp.JSONResponse) error {
	obj, err := json.Marshal(resp.Data)
	if err != nil {
		panic("invalid json")
	}
	_, err = rw.Write(obj)
	return err
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
	status safehttp.StatusCode
}

func newResponseRecorder(w io.Writer) *responseRecorder {
	return &responseRecorder{
		header: http.Header{},
		writer: w,
		status: http.StatusOK,
	}
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
