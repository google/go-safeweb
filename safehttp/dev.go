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

var (
	isLocalDev bool
	// freezeLocalDev is set on Mux construction.
	freezeLocalDev bool
)

// UseLocalDev instructs the framework to disable some security mechanisms that
// would make local development hard or impossible. This cannot be undone without
// restarting the program and should only be done before any other function or type
// of the framework is used.
// This function should ideally be called by the main package immediately after
// flag parsing.
// This configuration is not valid for production use.
func UseLocalDev() {
	if freezeLocalDev {
		panic("UseLocalDev can only be called before any other part of the framework")
	}
	isLocalDev = true
}

// IsLocalDev returns whether the framework is set up to use local development
// rules. Please see the doc on UseLocalDev.
func IsLocalDev() bool {
	return isLocalDev
}
