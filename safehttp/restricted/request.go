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

// Package restricted contains restricted APIs. Their usage should be security
// reviewed. You can use
// https://github.com/google/go-safeweb/tree/master/cmd/bancheck to enforce
// this.
package restricted

import (
	"net/http"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/internal"
)

var rawRequest = internal.RawRequest.(func(*safehttp.IncomingRequest) *http.Request)

// RawRequest returns the underlying *http.Request.
func RawRequest(r *safehttp.IncomingRequest) *http.Request {
	return rawRequest(r)
}
