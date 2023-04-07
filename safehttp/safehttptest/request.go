// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safehttptest

import (
	"io"
	"net/http/httptest"

	"github.com/google/go-safeweb/safehttp"
)

// NewRequest returns a new incoming server Request,
// suitable for passing to an http.Handler for testing.
//
// The target is the RFC 7230 "request-target": it may
// be either a path or an absolute URL. If target is an
// absolute URL, the host name from the URL is used.
// Otherwise, "example.com" is used.
//
// The TLS field is set to a non-nil dummy value if
// target has scheme "https".
//
// The Request.Proto is always HTTP/1.1.
//
// An empty method means "GET".
//
// The provided body may be nil. If the body is of type
// *bytes.Reader, *strings.Reader, or *bytes.Buffer, the
// Request.ContentLength is set.
//
// NewRequest panics on error for ease of use in testing,
// where a panic is acceptable.
//
// To generate a client HTTP request instead of a server
// request, see the NewRequest function in the net/http
// package.
func NewRequest(method, target string, body io.Reader) *safehttp.IncomingRequest {
	req := httptest.NewRequest(method, target, body)
	return safehttp.NewIncomingRequest(req)
}
