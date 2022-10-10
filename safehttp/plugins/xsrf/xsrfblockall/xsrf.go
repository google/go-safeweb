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

package xsrfblockall

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
)

// Interceptor implements XSRF protection.
type Interceptor struct{}

var _ safehttp.Interceptor = &Interceptor{}

// Before rejects every state changing request (all except GET, HEAD and OPTIONS).
func (it *Interceptor) Before(
	w safehttp.ResponseWriter,
	r *safehttp.IncomingRequest,
	_ safehttp.InterceptorConfig,
) safehttp.Result {
	if xsrf.StatePreserving(r) {
		return safehttp.NotWritten()
	}

	return w.WriteError(safehttp.StatusForbidden)
}

// Commit does nothing.
func (it *Interceptor) Commit(
	w safehttp.ResponseHeadersWriter,
	r *safehttp.IncomingRequest,
	resp safehttp.Response,
	_ safehttp.InterceptorConfig,
) {
}

// Match returns false since there are no supported configurations.
func (*Interceptor) Match(safehttp.InterceptorConfig) bool {
	return false
}
