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
	"net/http"
)

// RegisteredHandler returns the combined (all request methods) handler
// registered for a given pattern. Returns nil if the exact pattern wasn't used
// to register any handlers.
//
// This method is helpful for migrating services incrementally, endpoint by
// endpoint. The handler runs all the installed interceptors and the dispatcher.
//
// Important
//
// This function does not attempt to do any kind of path matching. If the
// handler was registered using the ServeMuxConfig for a pattern "/foo/", this
// method will return the handler only when given "/foo/" as an argument, not
// "/foo" nor "/foo/x".
func RegisteredHandler(mux *ServeMux, pattern string) http.Handler {
	if h, ok := mux.handlerMap[pattern]; ok {
		return h
	}
	// Keep this. Otherwise mux.handlerMap[pattern] returns a
	// (*registeredHandler)(nil), which is not equal to an untyped nil.
	return nil
}
