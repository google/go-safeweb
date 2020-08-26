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

package cors

import (
	"net/textproto"
	"strconv"
	"strings"

	"github.com/google/go-safeweb/safehttp"
)

const customHeader string = "X-Cors"

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
// - The request methods GET, HEAD and POST are banned and will result in a 405
//   Method Not Allowed response.
// - The content types "application/x-www-form-urlencoded", "multipart/form-data"
//   and "text/plain" are also banned and will result in a  415 Unsupported Media
//   Type response.
// - Each CORS request must contain the header "X-Cors: 1" or "Sec-Fetch-Mode: cors".
// All of this is to prevent XSRF.
type Interceptor struct {
	// AllowedOrigins determines which origins should be allowed in the
	// Access-Control-Allow-Origin header.
	AllowedOrigins map[string]bool
	allowedHeaders map[string]bool
	// ExposedHeaders determines which headers should be set in the
	// Access-Control-Expose-Headers header. This controls which headers in the
	// response should be accessible by JavaScript.
	//
	// If ExposedHeaders is nil, then the header is not set, meaning that nothing
	// is exposed.
	ExposedHeaders []string
	// AllowCredentials determines if Access-Control-Allow-Credentials should be
	// set to true, which would allow cookies to be attached to requests.
	AllowCredentials bool
	// MaxAge sets the Access-Control-Max-Age header.
	// MaxAge=0 results in MaxAge: 5.
	// This default of 5 seconds is set because it is the default used by Chromium
	// according to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
	MaxAge int
}

// Default creates a CORS Interceptor with default settings.
// Those defaults are:
//   - No Exposed Headers
//   - No Allowed Headers
//   - AllowCredentials: false
//   - MaxAge: 5 seconds
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

// SetAllowedHeaders sets which headers should be allowed in the Access-Control-Allow-Headers
// header. The headers are first canonicalized using textproto.CanonicalMIMEHeaderKey.
func (it *Interceptor) SetAllowedHeaders(headers ...string) {
	it.allowedHeaders = map[string]bool{}
	for _, h := range headers {
		it.allowedHeaders[textproto.CanonicalMIMEHeaderKey(h)] = true
	}
}

// Before handles the incoming request and sets headers or responds with an error
// accordingly.
func (it *Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	origin := r.Header.Get("Origin")
	if origin != "" && !it.AllowedOrigins[origin] {
		return w.ClientError(safehttp.StatusForbidden)
	}
	h := w.Header()
	allowOrigin, err := h.Claim("Access-Control-Allow-Origin")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	vary, err := h.Claim("Vary")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}

	allowCredentials, err := h.Claim("Access-Control-Allow-Credentials")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	if r.Header.Get("Cookie") != "" && !it.AllowCredentials {
		return w.ClientError(safehttp.StatusForbidden)
	}

	var status safehttp.StatusCode
	if r.Method() == safehttp.MethodOptions {
		status = it.preflight(w, r)
	} else {
		status = it.request(w, r)
	}

	if status != 0 && status != safehttp.StatusNoContent {
		if 400 <= status && status < 500 {
			return w.ClientError(status)
		} else {
			return w.ServerError(status)
		}
	}

	if origin != "" {
		allowOrigin([]string{origin})
		vary([]string{"Origin"})
	}
	if r.Header.Get("Cookie") != "" && it.AllowCredentials {
		// TODO: handle other credentials than cookies:
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
		allowCredentials([]string{"true"})
	}

	if status == safehttp.StatusNoContent {
		return w.NoContent()
	}
	return safehttp.Result{}
}

// preflight handles requests that have the method OPTIONS.
func (it *Interceptor) preflight(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.StatusCode {
	rh := r.Header
	method := rh.Get("Access-Control-Request-Method")
	if method == "" {
		return safehttp.StatusForbidden
	}
	wh := w.Header()
	allowMethods, err := wh.Claim("Access-Control-Allow-Methods")
	if err != nil {
		return safehttp.StatusInternalServerError
	}

	headers := rh.Get("Access-Control-Request-Headers")
	if headers != "" {
		for _, h := range strings.Split(headers, ", ") {
			h = textproto.CanonicalMIMEHeaderKey(h)
			if !it.allowedHeaders[h] && h != customHeader {
				return safehttp.StatusForbidden
			}
		}
	}
	allowHeaders, err := wh.Claim("Access-Control-Allow-Headers")
	if err != nil {
		return safehttp.StatusInternalServerError
	}

	maxAge, err := wh.Claim("Access-Control-Max-Age")
	if err != nil {
		return safehttp.StatusInternalServerError
	}

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
	if h.Get(customHeader) != "1" {
		return safehttp.StatusPreconditionFailed
	}

	if ct := h.Get("Content-Type"); ct == "" || disallowedContentTypes[ct] {
		return safehttp.StatusUnsupportedMediaType
	}

	exposeHeaders, err := w.Header().Claim("Access-Control-Expose-Headers")
	if err != nil {
		return safehttp.StatusInternalServerError
	}

	if it.ExposedHeaders != nil {
		exposeHeaders([]string{strings.Join(it.ExposedHeaders, ", ")})
	}
	return 0
}
