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

// Package cors provides a safehttp.Interceptor that handles CORS requests.
package cors

import (
	"net/textproto"
	"strconv"
	"strings"

	"github.com/google/go-safeweb/safehttp"
)

const requiredHeader string = "X-Cors"

var disallowedContentTypes = map[string]bool{
	"application/x-www-form-urlencoded": true,
	"multipart/form-data":               true,
	"text/plain":                        true,
}

// Interceptor handles CORS requests based on its settings.
//
// For more info about CORS, see: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
//
// Constraints
//
// The content types "application/x-www-form-urlencoded", "multipart/form-data"
// and "text/plain" are banned and will result in a  415 Unsupported Media Type
// response.
//
// Each CORS request must contain the header "X-Cors: 1".
//
// The HEAD request method is disallowed.
//
// All of this is to prevent XSRF.
type Interceptor struct {
	// AllowedOrigins determines which origins should be allowed in the
	// Access-Control-Allow-Origin header.
	AllowedOrigins map[string]bool
	// ExposedHeaders determines which headers should be set in the
	// Access-Control-Expose-Headers header. This controls which headers are
	//  accessible by JavaScript in the response.
	//
	// If ExposedHeaders is nil, then the header is not set, meaning that nothing
	// is exposed.
	ExposedHeaders []string
	// AllowCredentials determines if Access-Control-Allow-Credentials should be
	// set to true, which would allow cookies to be attached to requests.
	AllowCredentials bool
	// MaxAge sets the Access-Control-Max-Age header, indicating how many seconds
	// the results of a preflight request can be cached.
	//
	// MaxAge=0 results in MaxAge: 5, the default used by Chromium according to
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
	MaxAge         int
	allowedHeaders map[string]bool
}

var _ safehttp.Interceptor = &Interceptor{}

// Default creates a CORS Interceptor with default settings.
// Those defaults are:
//  - No Exposed Headers
//  - No Allowed Headers
//  - AllowCredentials: false
//  - MaxAge: 5 seconds
func Default(allowedOrigins ...string) *Interceptor {
	ao := map[string]bool{}
	for _, o := range allowedOrigins {
		ao[o] = true
	}
	return &Interceptor{
		AllowedOrigins: ao,
		allowedHeaders: map[string]bool{},
	}
}

// SetAllowedHeaders sets the headers allowed in the Access-Control-Allow-Headers
// header. The headers are first canonicalized using textproto.CanonicalMIMEHeaderKey.
// The wildcard "*" is not allowed.
func (it *Interceptor) SetAllowedHeaders(headers ...string) {
	it.allowedHeaders = map[string]bool{}
	for _, h := range headers {
		if h == "*" {
			continue
		}
		it.allowedHeaders[textproto.CanonicalMIMEHeaderKey(h)] = true
	}
}

// Before handles the IncomingRequest according to the settings specified in the
// Interceptor and sets the appropriate subset of the following headers:
//
//  - Access-Control-Allow-Credentials
//  - Access-Control-Allow-Headers
//  - Access-Control-Allow-Methods
//  - Access-Control-Allow-Origin
//  - Access-Control-Expose-Headers
//  - Access-Control-Max-Age
//  - Vary
func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, _ safehttp.InterceptorConfig) safehttp.Result {
	origin := r.Header.Get("Origin")
	if origin != "" && !it.AllowedOrigins[origin] {
		return w.WriteError(safehttp.StatusForbidden)
	}
	h := w.Header()
	allowOrigin := h.Claim("Access-Control-Allow-Origin")
	if h.IsClaimed("Vary") {
		return w.WriteError(safehttp.StatusInternalServerError)
	}

	allowCredentials := h.Claim("Access-Control-Allow-Credentials")

	var status safehttp.StatusCode
	switch r.Method() {
	case safehttp.MethodOptions:
		status = it.preflight(w, r)
	case safehttp.MethodHead:
		status = safehttp.StatusMethodNotAllowed
	default:
		status = it.request(w, r)
	}

	if status != 0 && status != safehttp.StatusNoContent {
		return w.WriteError(status)
	}

	if origin != "" {
		allowOrigin([]string{origin})
		appendToVary(w, "Origin")
	}
	if r.Header.Get("Cookie") != "" && it.AllowCredentials {
		// TODO: handle other credentials than cookies:
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
		allowCredentials([]string{"true"})
	}

	if status == safehttp.StatusNoContent {
		return w.NoContent()
	}
	return safehttp.NotWritten()
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (it *Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) safehttp.Result {
	return safehttp.NotWritten()
}

func appendToVary(w *safehttp.ResponseWriter, val string) {
	h := w.Header()
	if curr := h.Get("Vary"); curr != "" {
		h.Set("Vary", curr+", "+val)
	} else {
		h.Set("Vary", val)
	}
}

// preflight handles requests that have the method OPTIONS.
func (it *Interceptor) preflight(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.StatusCode {
	rh := r.Header
	if rh.Get("Origin") == "" {
		return safehttp.StatusForbidden
	}

	method := rh.Get("Access-Control-Request-Method")
	if method == "" || method == safehttp.MethodHead {
		return safehttp.StatusForbidden
	}

	headers := rh.Get("Access-Control-Request-Headers")
	if headers != "" {
		for _, h := range strings.Split(headers, ", ") {
			h = textproto.CanonicalMIMEHeaderKey(h)
			if !it.allowedHeaders[h] && h != requiredHeader {
				return safehttp.StatusForbidden
			}
		}
	}

	wh := w.Header()
	allowMethods := wh.Claim("Access-Control-Allow-Methods")
	allowHeaders := wh.Claim("Access-Control-Allow-Headers")
	maxAge := wh.Claim("Access-Control-Max-Age")

	allowMethods([]string{method})
	if headers != "" {
		allowHeaders([]string{headers})
	}
	n := it.MaxAge
	if n == 0 {
		n = 5
	}
	maxAge([]string{strconv.Itoa(n)})
	return safehttp.StatusNoContent
}

// request handles all requests that are not preflight requests.
func (it *Interceptor) request(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.StatusCode {
	h := r.Header
	if h.Get(requiredHeader) != "1" {
		return safehttp.StatusPreconditionFailed
	}

	if ct := h.Get("Content-Type"); ct == "" || disallowedContentTypes[ct] {
		return safehttp.StatusUnsupportedMediaType
	}

	if exposeHeaders := w.Header().Claim("Access-Control-Expose-Headers"); it.ExposedHeaders != nil {
		exposeHeaders([]string{strings.Join(it.ExposedHeaders, ", ")})
	}
	return 0
}
