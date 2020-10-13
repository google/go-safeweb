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

// Write writes the response to the http.ResponseWriter if it's deemed safe.
// A safe response is either safe HTML, a JSON object or safe HTML template. It
// will  return a non-nil error if the response is deemed unsafe or if writing
// operation fails.
func (DefaultDispatcher) Write(rw http.ResponseWriter, resp Response) error {
	switch x := resp.(type) {
	case JSONResponse:
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(int(StatusOK))
		io.WriteString(rw, ")]}',\n") // Break parsing of JavaScript in order to prevent XSSI.
		return json.NewEncoder(rw).Encode(x.Data)
	case TemplateResponse:
		t, ok := (*x.Template).(*template.Template)
		if !ok {
			return fmt.Errorf("%T is not a safe template and it cannot be parsed and written", t)
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(int(StatusOK))
		if len(x.FuncMap) == 0 {
			return t.Execute(rw, x.Data)
		}
		cloned, err := t.Clone()
		if err != nil {
			return err
		}
		return cloned.Funcs(x.FuncMap).Execute(rw, x.Data)
	default:
		r, ok := x.(safehtml.HTML)
		if !ok {
			return fmt.Errorf("%T is not a safe response type and it cannot be written", resp)
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(int(StatusOK))
		_, err := io.WriteString(rw, r.String())
		return err
	}
}
