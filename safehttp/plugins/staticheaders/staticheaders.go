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

// Package staticheaders provides a safehttp.Interceptor which sets security
// sensitive headers on every response.
//
// X-Content-Type-Options: nosniff - tells browsers to not to sniff the
// Content-Type of responses (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options).
//
// X-XSS-Protection: 0 - tells the browser to disable any built in XSS filters.
// These built in XSS filters are unnecessary when other, stronger, protections
// are available and can introduce cross-site leaks vulnerabilities
// (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection).
//
// Usage
//
// Install an instance of Interceptor using safehttp.ServerMux.Install.
package staticheaders

import (
	"github.com/google/go-safeweb/safehttp"
)

// Interceptor claims and sets static headers on responses.
// The zero value is valid and ready to use.
type Interceptor struct{}

var _ safehttp.Interceptor = Interceptor{}

// Before claims and sets the following headers:
//  - X-Content-Type-Options: nosniff
//  - X-XSS-Protection: 0
func (Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	h := w.Header()
	setXCTO := h.Claim("X-Content-Type-Options")
	setXXP := h.Claim("X-XSS-Protection")

	setXCTO([]string{"nosniff"})
	setXXP([]string{"0"})
	return safehttp.NotWritten()
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}

// Match returns false since there are no supported configurations.
func (Interceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}
