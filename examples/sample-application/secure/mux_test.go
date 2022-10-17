// Copyright 2022 Google LLC
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

package secure

import (
	"net/http/httptest"
	"testing"

	"github.com/google/go-safeweb/safehttp"
)

func TestNewMuxConfig(t *testing.T) {
	mux := NewMuxConfig(nil, "foo:8080").Mux()
	mux.Handle("/", safehttp.MethodGet, nil)
	req := httptest.NewRequest(safehttp.MethodGet, "http://foo:8080/", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)
}
