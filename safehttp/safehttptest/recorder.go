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

package safehttptest

import (
	"net/http"
	"net/http/httptest"

	"github.com/google/go-safeweb/safehttp"
)

// ResponseRecorder encapsulates a safehttp.ResponseWriter that records
// mutations for later inspection in tests. The safehttp.ResponseWriter
// should be passed as part of the handler function in tests.
type ResponseRecorder struct {
	safehttp.ResponseWriter
	rw *httptest.ResponseRecorder
}

// NewResponseRecorder creates a ResponseRecorder.
func NewResponseRecorder() *ResponseRecorder {
	rw := httptest.NewRecorder()
	return &ResponseRecorder{
		rw:             rw,
		ResponseWriter: safehttp.DeprecatedNewResponseWriter(rw, nil),
	}
}

// NewResponseRecorderFromDispatcher creates a ResponseRecorder.
func NewResponseRecorderFromDispatcher(d safehttp.Dispatcher) *ResponseRecorder {
	rw := httptest.NewRecorder()
	return &ResponseRecorder{
		rw:             rw,
		ResponseWriter: safehttp.DeprecatedNewResponseWriter(rw, d),
	}
}

// Header returns the recorded response headers.
func (r *ResponseRecorder) Header() http.Header {
	return r.rw.Header()
}

// Status returns the recorded response status code.
func (r *ResponseRecorder) Status() safehttp.StatusCode {
	return safehttp.StatusCode(r.rw.Code)
}

// Body returns the recorded response body.
func (r *ResponseRecorder) Body() string {
	return r.rw.Body.String()
}
