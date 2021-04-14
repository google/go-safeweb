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
	"io"
	"net/http"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

// DefaultDispatcher is responsible for writing safe responses.
type DefaultDispatcher struct{}

// ContentType returns the Content-Type of a response if it's deemed safe and
// an error otherwise.
func (DefaultDispatcher) ContentType(resp Response) (string, error) {
	switch x := resp.(type) {
	case safehtml.HTML:
		return "text/html; charset=utf-8", nil
	case *TemplateResponse:
		_, ok := (x.Template).(*template.Template)
		if !ok {
			return "", fmt.Errorf("%T is not a safe response type, a Content-Type cannot be provided", resp)
		}
		return "text/html; charset=utf-8", nil
	case JSONResponse:
		return "application/json; charset=utf-8", nil
	case FileServerResponse:
		return x.ContentType(), nil
	default:
		return "", fmt.Errorf("%T is not a safe response type, a Content-Type cannot be provided", resp)
	}
}

// Write writes the response to the http.ResponseWriter if it's deemed safe. It
// returns a non-nil error if the response is deemed unsafe or if the writing
// operation fails.
//
// For JSONResponses, the underlying object is serialised and written if it's a
// valid JSON.
//
// For TemplateResponses, the parsed template is applied to the provided data
// object. If the funcMap is non-nil, its elements override the  existing names
// to functions mappings in the template. An attempt to define a new name to
// function mapping that is not already in the template will result  in a panic.
func (DefaultDispatcher) Write(rw http.ResponseWriter, resp Response) error {
	switch x := resp.(type) {
	case JSONResponse:
		io.WriteString(rw, ")]}',\n") // Break parsing of JavaScript in order to prevent XSSI.
		return json.NewEncoder(rw).Encode(x.Data)
	case *TemplateResponse:
		t, ok := (x.Template).(*template.Template)
		if !ok {
			return fmt.Errorf("%T is not a safe template and it cannot be parsed and written", t)
		}
		if len(x.FuncMap) == 0 {
			if x.Name == "" {
				return t.Execute(rw, x.Data)
			} else {
				return t.ExecuteTemplate(rw, x.Name, x.Data)
			}
		}
		cloned, err := t.Clone()
		if err != nil {
			return err
		}
		cloned = cloned.Funcs(x.FuncMap)
		if x.Name == "" {
			return cloned.Execute(rw, x.Data)
		} else {
			return cloned.ExecuteTemplate(rw, x.Name, x.Data)
		}
	case safehtml.HTML:
		_, err := io.WriteString(rw, x.String())
		return err
	case FileServerResponse:
		// The http package will take care of writing the file body.
		return nil
	default:
		return fmt.Errorf("%T is not a safe response type and it cannot be written", resp)
	}
}

// Error writes the error response to the http.ResponseWriter if it's deemed
// safe. It returns a non-nil error if the response is deemed unsafe or if the
// writing operation fails.
func (DefaultDispatcher) Error(rw http.ResponseWriter, resp ErrorResponse) error {
	switch x := resp.(type) {
	case StatusCode:
		http.Error(rw, http.StatusText(int(x.Code())), int(x.Code()))
		return nil
	default:
		return fmt.Errorf("%T is not a safe response type and it cannot be written", resp)
	}
}
