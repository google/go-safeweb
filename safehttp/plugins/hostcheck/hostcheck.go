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

// Package hostcheck provides a plugin that checks whether the request is
// intended to be sent to a given host.
//
// This is a protection mechanism against
// DNS rebinding attacks (https://en.wikipedia.org/wiki/DNS_rebinding) and HTTP
// request smuggling (https://portswigger.net/web-security/request-smuggling).
package hostcheck

import (
	"github.com/google/go-safeweb/safehttp"
)

// Interceptor checks whether the Host header of the incoming request is in an
// allowlist.
type Interceptor struct {
	hosts map[string]bool
}

var _ safehttp.Interceptor = Interceptor{}

// New creates an Interceptor.
func New(hosts ...string) Interceptor {
	it := Interceptor{hosts: map[string]bool{}}
	for _, h := range hosts {
		it.hosts[h] = true
	}
	return it
}

// Before checks whether the request's Host header is in the list of allowed
// hosts. If it's not, it responds with 404 Not Found.
func (it Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	if !it.hosts[r.Host()] {
		return w.WriteError(safehttp.StatusNotFound)
	}
	return safehttp.NotWritten()
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}
