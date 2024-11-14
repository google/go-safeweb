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

// Package framing provides utilities to install a comprehensive framing protection.
package framing

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp"

	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata"
	"github.com/google/go-safeweb/safehttp/plugins/framing/internalunsafeframing"
)

// Interceptors returns all interceptors needed for a comprehensive framing protection.
func Interceptors(reportURI string) []safehttp.Interceptor {
	return []safehttp.Interceptor{
		fetchmetadata.FramingIsolationPolicy(),
		csp.Interceptor{Policy: csp.FramingPolicy{ReportURI: reportURI}},
		xfoInterceptor{},
	}
}

type xfoInterceptor struct{}

func (xfoInterceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	xfo := w.Header().Claim("X-Frame-Options")
	switch cfg.(type) {
	case internalunsafeframing.Disable, internalunsafeframing.AllowList:
		// X-Frame-Options doesn't support allowlists.
		// We rely on CSP to do the restriction on this value.
		xfo([]string{"ALLOWALL"})
	default:
		xfo([]string{"SAMEORIGIN"})
	}
	return safehttp.NotWritten()
}

func (xfoInterceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
}

func (xfoInterceptor) Match(cfg safehttp.InterceptorConfig) bool {
	switch cfg.(type) {
	case internalunsafeframing.Disable, internalunsafeframing.AllowList:
		return true
	}
	return false
}
