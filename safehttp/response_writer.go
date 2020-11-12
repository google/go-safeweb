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

// ResponseWriter is used to construct an HTTP response. When a Response is
// passed to the ResponseWriter, it will invoke the Dispatcher with the
// Response. An attempt to write to the ResponseWriter twice will
// cause a panic.
//
// A ResponseWriter may not be used after the Handler.ServeHTTP method has returned.
type ResponseWriter interface {
	// Write writes a safe response.
	Write(resp Response) Result

	// NoContent responds with a 204 No Content response.
	//
	// If the ResponseWriter has already been written to, then this method panics.
	NoContent() Result

	// WriteError writes an error response (400-599) according to the provided status code.
	//
	// If the ResponseWriter has already been written to, then this method panics.
	WriteError(code StatusCode) Result

	// Redirect responds with a redirect to the given url, using code as the status code.
	//
	// If the ResponseWriter has already been written to, then this method panics.
	Redirect(r *IncomingRequest, url string, code StatusCode) Result

	// Header returns the collection of headers that will be set
	// on the response. Headers must be set before writing a
	// response (e.g. Write, WriteTemplate).
	Header() Header

	// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
	// The provided cookie must have a valid Name, otherwise an error will be
	// returned.
	SetCookie(c *Cookie) error

	// SetCode allows setting a response status. If the response was already
	// written, trying to set the status code will have no effect. This method will
	// panic if an invalid status code is passed (i.e. not in the range 1XX-5XX).
	//
	// If SetCode was called before NoContent, Redirect or WriteError, the status
	// code set by the latter will be the actual response status.
	SetCode(code StatusCode)
}
