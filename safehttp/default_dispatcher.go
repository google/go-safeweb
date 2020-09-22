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
	"fmt"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
	"io"
	"net/http"
)

// DefaultDispatcher is responsible for writing safe responses.
type DefaultDispatcher struct{}

// ContentType returns the Content-Type of a safe response if it's of a
// safe-type and an error otherwise.
func (DefaultDispatcher) ContentType(resp Response) (string, error) {
	switch resp.(type) {
	case safehtml.HTML, *template.Template:
		return "text/html; charset=utf-8", nil
	case JSONResponse:
		return "application/json; charset=utf-8", nil
	default:
		return "", fmt.Errorf("%T is not a safe response type, a Content-Type cannot be provided", resp)
	}
}

// Write writes the response to the Response Writer if it's safe HTML. It
// returns a non-nil error if the response is not safe HTML or if writing the
// response fails.
func (DefaultDispatcher) Write(rw http.ResponseWriter, resp Response) error {
	x, ok := resp.(safehtml.HTML)
	if !ok {
		return fmt.Errorf("%T is not a safe response type and it cannot be written", resp)
	}
	_, err := io.WriteString(rw, x.String())
	return err

}

// WriteJSON serialises and writes a JSON to the underlying http.ResponseWriter.
// If the provided response is not valid JSON or writing the response fails, the
// method will return an error.
func (DefaultDispatcher) WriteJSON(rw http.ResponseWriter, resp JSONResponse) error {
	io.WriteString(rw, ")]}',\n") // Break parsing of JavaScript in order to prevent XSSI.
	return json.NewEncoder(rw).Encode(resp.Data)
}

// ExecuteTemplate applies the parsed template to the provided data object, if
// the template is a safe HTML template, writing the output to the  http.
// ResponseWriter. If the funcMap is non-nil, its elements can override the
// existing names to functions mappings in the template. An attempt to define a
// new name to function mapping that is not already in the template will result
// in a panic. The template, data object and funcMap are contained in the
// TemplateResponse.
//
// If an error occurs executing the template or writing its output,
// execution stops, but partial results may already have been written to
// the Response Writer.
func (DefaultDispatcher) ExecuteTemplate(rw http.ResponseWriter, resp TemplateResponse) error {
	t := *resp.Template
	x, ok := t.(*template.Template)
	if !ok {
		return fmt.Errorf("%T is not a safe template and it cannot be parsed and written", resp.Template)
	}
	if len(resp.FuncMap) == 0 {
		return x.Execute(rw, resp.Data)
	}
	cloned, err := x.Clone()
	if err != nil {
		return err
	}
	return cloned.Funcs(resp.FuncMap).Execute(rw, resp.Data)
}
