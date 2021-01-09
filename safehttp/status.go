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

import "net/http"

// StatusCode contains HTTP status codes as registered with IANA.
// See: https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
type StatusCode int

// The HTTP status codes registered with IANA.
const (
	StatusContinue           StatusCode = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols StatusCode = 101 // RFC 7231, 6.2.2
	StatusProcessing         StatusCode = 102 // RFC 2518, 10.1
	StatusEarlyHints         StatusCode = 103 // RFC 8297

	StatusOK                   StatusCode = 200 // RFC 7231, 6.3.1
	StatusCreated              StatusCode = 201 // RFC 7231, 6.3.2
	StatusAccepted             StatusCode = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo StatusCode = 203 // RFC 7231, 6.3.4
	StatusNoContent            StatusCode = 204 // RFC 7231, 6.3.5
	StatusResetContent         StatusCode = 205 // RFC 7231, 6.3.6
	StatusPartialContent       StatusCode = 206 // RFC 7233, 4.1
	StatusMultiStatus          StatusCode = 207 // RFC 4918, 11.1
	StatusAlreadyReported      StatusCode = 208 // RFC 5842, 7.1
	StatusIMUsed               StatusCode = 226 // RFC 3229, 10.4.1

	StatusMultipleChoice   StatusCode = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently StatusCode = 301 // RFC 7231, 6.4.2
	StatusFound            StatusCode = 302 // RFC 7231, 6.4.3
	StatusSeeOther         StatusCode = 303 // RFC 7231, 6.4.4
	StatusNotModified      StatusCode = 304 // RFC 7232, 4.1
	StatusUseProxy         StatusCode = 305 // RFC 7231, 6.4.5

	StatusTemporaryRedirect StatusCode = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect StatusCode = 308 // RFC 7538, 3

	StatusBadRequest                   StatusCode = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 StatusCode = 401 // RFC 7235, 3.1
	StatusPaymentRequired              StatusCode = 402 // RFC 7231, 6.5.2
	StatusForbidden                    StatusCode = 403 // RFC 7231, 6.5.3
	StatusNotFound                     StatusCode = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed             StatusCode = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                StatusCode = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired            StatusCode = 407 // RFC 7235, 3.2
	StatusRequestTimeout               StatusCode = 408 // RFC 7231, 6.5.7
	StatusConflict                     StatusCode = 409 // RFC 7231, 6.5.8
	StatusGone                         StatusCode = 410 // RFC 7231, 6.5.9
	StatusLengthRequired               StatusCode = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed           StatusCode = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge        StatusCode = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong            StatusCode = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType         StatusCode = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable StatusCode = 416 // RFC 7233, 4.4
	StatusExpectationFailed            StatusCode = 417 // RFC 7231, 6.5.14
	StatusTeapot                       StatusCode = 418 // RFC 7168, 2.3.3
	StatusMisdirectedRequest           StatusCode = 421 // RFC 7540, 9.1.2
	StatusUnprocessableEntity          StatusCode = 422 // RFC 4918, 11.2
	StatusLocked                       StatusCode = 423 // RFC 4918, 11.3
	StatusFailedDependency             StatusCode = 424 // RFC 4918, 11.4
	StatusTooEarly                     StatusCode = 425 // RFC 8470, 5.2.
	StatusUpgradeRequired              StatusCode = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         StatusCode = 428 // RFC 6585, 3
	StatusTooManyRequests              StatusCode = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarg   StatusCode = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   StatusCode = 451 // RFC 7725, 3

	StatusInternalServerError           StatusCode = 500 // RFC 7231, 6.6.1
	StatusNotImplemented                StatusCode = 501 // RFC 7231, 6.6.2
	StatusBadGateway                    StatusCode = 502 // RFC 7231, 6.6.3
	StatusServiceUnavailable            StatusCode = 503 // RFC 7231, 6.6.4
	StatusGatewayTimeout                StatusCode = 504 // RFC 7231, 6.6.5
	StatusHTTPVersionNotSupported       StatusCode = 505 // RFC 7231, 6.6.6
	StatusVariantAlsoNegotiates         StatusCode = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           StatusCode = 507 // RFC 4918, 11.5
	StatusLoopDetected                  StatusCode = 508 // RFC 5842, 7.2
	StatusNotExtended                   StatusCode = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired StatusCode = 511 // RFC 6585, 6
)

// Code implements ErrorResponse.
func (c StatusCode) Code() StatusCode {
	return c
}

// String returns the status text.
func (c StatusCode) String() string {
	return http.StatusText(int(c))
}

// Write writes the error response.
func (c StatusCode) Write(rw http.ResponseWriter) {
	http.Error(rw, c.String(), int(c.Code()))
}
