// Copyright 2022 Google LLC
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

package secure

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/examples/sample-application/secure/responses"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

func TestDispatcherError(t *testing.T) {
	tests := []struct {
		name string
		resp safehttp.ErrorResponse
		want string
	}{
		{
			name: "default dispatcher, plaintext",
			resp: safehttp.StatusForbidden,
			want: http.StatusText(int(safehttp.StatusForbidden)),
		},
		{
			name: "custom dispatcher, safe HTML",
			resp: responses.Error{
				StatusCode: safehttp.StatusNotFound,
				Message:    safehtml.HTMLEscaped("<h1>Escaped</h1"),
			},
			want: "<p>&lt;h1&gt;Escaped&lt;/h1</p>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dispatcher{}
			rw := httptest.NewRecorder()
			d.Error(rw, tt.resp)
			if op := rw.Body.String(); !strings.Contains(op, tt.want) {
				t.Errorf("body got: %q, want: %q", op, tt.want)
			}
		})
	}
}
