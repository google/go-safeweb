// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package xsrf contains helper functions for the safehttp.Interceptor that
// provide protection against Cross-Site Request Forgery attacks.
package xsrf

import (
	"github.com/google/go-safeweb/safehttp"
)

var statePreservingMethods = map[string]bool{
	safehttp.MethodGet:     true,
	safehttp.MethodHead:    true,
	safehttp.MethodOptions: true,
}

// StatePreserving checks if the provided request is state preserving.
func StatePreserving(r *safehttp.IncomingRequest) bool {
	return statePreservingMethods[r.Method()]
}
