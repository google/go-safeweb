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

// Handler responds to an HTTP request.
//
// ServeHTTP should set headers, call one of the writing functions on the ResponseWriter
// to write the body of the response and then return the Result received from writing
// to the ResponseWriter.
//
// It is not valid to use the ResponseWriter or read from the body of the IncomingRequest
// after or concurrently with the completion of the ServeHTTP call.
//
// Except for reading the body, handlers should not modify the provided Request.
//
// TODO: Add documentation about error handling when properly implemented.
type Handler interface {
	ServeHTTP(*ResponseWriter, *IncomingRequest) Result
}

// HandlerFunc can be used to convert a function into a Handler without creating
// a new struct. If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler that calls f.
type HandlerFunc func(*ResponseWriter, *IncomingRequest) Result

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w *ResponseWriter, r *IncomingRequest) Result {
	return f(w, r)
}
