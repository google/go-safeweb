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

import "net/http"

// Dispatcher is responsible for writing a response received from the
// ResponseWriter to the underlying http.ResponseWriter.
//
// The implementation of a custom Dispatcher should be thoroughly reviewed by
// the security team to avoid introducing vulnerabilities.
type Dispatcher interface {
	// Write writes a Response to the underlying http.ResponseWriter.
	//
	// Write is responsible for setting the Content-Type response header. If the
	// Dispatcher doesn't set the HTTP response status code, the default
	// behavior of http.ResponseWriter applies (i.e. 200 OK is set on first
	// Write).
	//
	// It should return an error if the writing operation fails or if the
	// provided Response should not be written to the http.ResponseWriter
	// because it's unsafe.
	Write(rw http.ResponseWriter, resp Response) error

	// Error writes an ErrorResponse to the underlying http.ResponseWriter.
	//
	// Error is responsible for setting the Content-Type response header and the
	// HTTP response status code.
	//
	// It should return an error if the writing operation fails.
	//
	// Error should always attempt to write a response, no matter what is the
	// underlying type of resp. As a fallback, the Dispatcher can use WriteTextError.
	Error(rw http.ResponseWriter, resp ErrorResponse) error
}

func writeTextError(rw http.ResponseWriter, resp ErrorResponse) {
	http.Error(rw, http.StatusText(int(resp.Code())), int(resp.Code()))
}
