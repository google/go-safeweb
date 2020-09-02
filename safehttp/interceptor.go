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

// Interceptor can be installed on a ServeMux in order to apply its
// functionality on an IncomingRequest before it is sent to its corresponding
// handler.
type Interceptor interface {
	// Before runs before the IncomingRequest is sent to the handler. If a
	// response is written to the ResponseWriter, then the remaining
	// interceptors and the handler won't execute. If Before panics, it will be
	// recovered and the ServeMux will respond with 500 Internal Server Error.
	Before(w *ResponseWriter, r *IncomingRequest, cfg interface{}) Result
}
