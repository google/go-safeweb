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

package reportscollector

import (
	"encoding/json"

	"github.com/google/go-safeweb/safehttp"
)

// Report represents a CSP violation report as specified by
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy-Report-Only#Violation_report_syntax
type Report struct {
	BlockedURI         string `json:"blocked-uri"`
	Disposition        string `json:"disposition"`
	DocumentURI        string `json:"document-uri"`
	EffectiveDirective string `json:"effective-directive"`
	OriginalPolicy     string `json:"original-policy"`
	Referrer           string `json:"referrer"`
	ScriptSample       string `json:"script-sample"`
	StatusCode         int    `json:"status-code"`
	ViolatedDirective  string `json:"violated-directive"`
}

type report struct{}

// Collector is a reports collector.
type Collector interface {
	Collect(Report)
}

// Handler creates a safehttp.Handler which calls collector when a violation
// report is received. Make sure to register the handler to receive POST
// requests. If the handler recieves anything other that POST requests it will
// respond with a 405 Method Not Allowed.
func Handler(c Collector) safehttp.Handler {
	return safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
		if ir.Method() != safehttp.MethodPost {
			return w.ClientError(safehttp.StatusMethodNotAllowed)
		}

		if ir.Header.Get("Content-Type") != "application/json" {
			return w.ClientError(safehttp.StatusUnsupportedMediaType)
		}

		d := json.NewDecoder(ir.Body())
		r := Report{}
		if err := d.Decode(&r); err != nil {
			return w.ClientError(safehttp.StatusBadRequest)
		}
		c.Collect(r)

		return w.NoContent()
	})
}
