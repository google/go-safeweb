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

package collector

import (
	"encoding/json"
	"io/ioutil"

	"github.com/google/go-safeweb/safehttp"
)

// Report represents a CSP violation report as specified by
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy-Report-Only#Violation_report_syntax
type Report struct {
	// BlockedURI is the URI of the resource that was blocked from loading by the
	// Content Security Policy. If the blocked URI is from a different origin than
	// the DocumentURI, then the blocked URI is truncated to contain just the scheme,
	// host, and port.
	BlockedURI string `json:"blocked-uri"`
	// Disposition is either "enforce" or "report" depending on whether the Content-Security-Policy
	// header or the Content-Security-Policy-Report-Only header is used.
	Disposition string `json:"disposition"`
	// DocumentURI is the URI of the document in which the violation occurred.
	DocumentURI string `json:"document-uri"`
	// EffectiveDirective is the directive whose enforcement caused the violation.
	EffectiveDirective string `json:"effective-directive"`
	// OriginalPolicy is the original policy as specified by the Content Security
	// Policy header.
	OriginalPolicy string `json:"original-policy"`
	// Referrer is the referrer of the document in which the violation occurred.
	Referrer string `json:"referrer"`
	// ScriptSample is the first 40 characters of the inline script, event handler,
	// or style that caused the violation.
	ScriptSample string `json:"script-sample"`
	// StatusCode the HTTP status code of the resource on which the global object
	// was instantiated.
	StatusCode int `json:"status-code"`
	// ViolatedDirective is the name of the policy section that was violated.
	ViolatedDirective string `json:"violated-directive"`
}

// ReportsHandler is a reports handler.
type ReportsHandler func(Report)

// Handler creates a safehttp.Handler which calls the given ReportsHandler when a
// violation report is received. Make sure to register the handler to receive POST
// requests. If the handler recieves anything other that POST requests it will
// respond with a 405 Method Not Allowed.
func Handler(rh ReportsHandler) safehttp.Handler {
	return safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
		if ir.Method() != safehttp.MethodPost {
			return w.ClientError(safehttp.StatusMethodNotAllowed)
		}

		if ir.Header.Get("Content-Type") != "application/json" {
			return w.ClientError(safehttp.StatusUnsupportedMediaType)
		}

		b, err := ioutil.ReadAll(ir.Body())
		if err != nil {
			return w.ClientError(safehttp.StatusBadRequest)
		}
		r := Report{}
		if err := json.Unmarshal(b, &r); err != nil {
			return w.ClientError(safehttp.StatusBadRequest)
		}
		rh(r)

		return w.NoContent()
	})
}
