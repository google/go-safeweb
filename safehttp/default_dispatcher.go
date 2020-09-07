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
	"encoding/json"
	"errors"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
	"net/http"
)

// DefaultDispatcher is responsible of writing the body of a response to an
// Response Writer if it's of a safe type.
type DefaultDispatcher struct{}

// ContentType returns the Content-Type of a safe respons if it's of a safe-type
// and a non-nil error otherwise.
func (DefaultDispatcher) ContentType(resp Response) (string, error) {
	switch resp.(type) {
	case safehtml.HTML, *template.Template:
		return "text/html; charset=utf-8", nil
	case JSONResponse:
		return "application/json; charset=utf-8", nil
	default:
		return "", errors.New("not a safe response")
	}
}

// Write writes the response to the Response Writer if it's safe HTML. It
// returns a non-nil error if the response is not safe HTML or if writing the
// response fails.
func (DefaultDispatcher) Write(rw http.ResponseWriter, resp Response) error {
	x, ok := resp.(safehtml.HTML)
	if !ok {
		return errors.New("not a safe response type")
	}
	_, err := rw.Write([]byte(x.String()))
	return err

}

// WriteJSON serialises and writes a JSON to the Response Writer. If the
// provided response is not valid JSON or writing the response fails, the method
// will return an error.
func (DefaultDispatcher) WriteJSON(rw http.ResponseWriter, resp JSONResponse) error {
	obj, err := json.Marshal(resp.Data)
	if err != nil {
		return errors.New("invalid json")
	}
	_, err = rw.Write(obj)
	return err
}

// ExecuteTemplate applies a parsed template to the provided data object if the
// template is a safe HTML template, writing the output to the ResponseWriter.
//
// If an error occurs executing the template or writing its output,
// execution stops, but partial results may already have been written to
// the Response Writer.
func (DefaultDispatcher) ExecuteTemplate(rw http.ResponseWriter, t Template, data interface{}) error {
	x, ok := t.(*template.Template)
	if !ok {
		return errors.New("not a safe response type")
	}
	err := x.Execute(rw, data)
	return err
}
