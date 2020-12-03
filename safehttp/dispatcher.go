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
	// Content-Type returns the Content-Type of the provided response if it is
	// of a safe type, supported by the Dispatcher, and should return an error
	// otherwise.
	//
	// Sending a response to the http.ResponseWriter without properly setting
	// CT is error-prone and could introduce vulnerabilities. Therefore, this
	// method should be used to set the Content-Type header before calling
	// Dispatcher.Write. Writing should not proceed if ContentType returns an
	// error.
	ContentType(resp Response) (string, error)

	// Write writes a Response to the underlying http.ResponseWriter.
	//
	// It should return an error if the writing operation fails or if the
	// provided Response should not be written to the http.ResponseWriter
	// because it's unsafe.
	Write(rw http.ResponseWriter, resp Response) error

	// Error writes an ErrorResponse to the underlying http.ResponseWriter.
	//
	// It should return an error if the writing operation fails or if the
	// provided Response should not be written to the http.ResponseWriter
	// because it's unsafe.
	Error(rw http.ResponseWriter, resp ErrorResponse) error
}
