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

// Package collector provides a function for creating violation report handlers.
// The created safehttp.Handler will be able to parse generic violation reports
// as specified by https://w3c.github.io/reporting/ and CSP violation reports as
// specified by https://www.w3.org/TR/CSP3/#deprecated-serialize-violation.
package collector

import (
	"encoding/json"
	"io/ioutil"

	"github.com/google/go-safeweb/safehttp"
)

// Report represents a generic report as specified by https://w3c.github.io/reporting/#serialize-reports
type Report struct {
	// Type represents the type of the report. This will control how Body looks
	// like.
	Type string
	// Age represents the number of milliseconds since the violation causing the
	// report occured.
	Age uint64
	// URL is the address of the Document or Worker from which the report was
	// generated.
	URL string
	// UserAgent contains the value of the User-Agent header of the request from
	// which the report was generated.
	UserAgent string
	// Body contains the body of the report. This will be different for every Type.
	// If Type is csp-violation then Body will be a CSPReport. Otherwise Body will
	// be a map[string]interface{} containing the object that was passed, as unmarshalled
	// using encoding/json.
	Body interface{}
}

// CSPReport represents a CSP violation report as specified by https://www.w3.org/TR/CSP3/#deprecated-serialize-violation
type CSPReport struct {
	// BlockedURL is the URL of the resource that was blocked from loading by the
	// Content Security Policy. If the blocked URL is from a different origin than
	// the DocumentURL, the blocked URL is truncated to contain just the scheme,
	// host and port.
	BlockedURL string
	// Disposition is either "enforce" or "report" depending on whether the Content-Security-Policy
	// header or the Content-Security-Policy-Report-Only header is used.
	Disposition string
	// DocumentURL is the URL of the document in which the violation occurred.
	DocumentURL string
	// EffectiveDirective is the directive whose enforcement caused the violation.
	EffectiveDirective string
	// OriginalPolicy is the original policy as specified by the Content Security
	// Policy header.
	OriginalPolicy string
	// Referrer is the referrer of the document in which the violation occurred.
	Referrer string
	// Sample is the first 40 characters of the inline script, event handler,
	// or style that caused the violation.
	Sample string
	// StatusCode is the HTTP status code of the resource on which the global object
	// was instantiated.
	StatusCode uint
	// ViolatedDirective is the name of the policy section that was violated.
	ViolatedDirective string
	// SourceFile represents the URL of the document or worker in which the violation
	// was found.
	SourceFile string
	// LineNumber is the line number in the document or worker at which the violation
	// occurred.
	LineNumber uint
	// ColumnNumber is the column number in the document or worker at which the violation
	// occurred.
	ColumnNumber uint
}

// Handler builds a safehttp.Handler which calls the given handler or cspHandler when
// a violation report is received. Make sure to register the handler to receive POST
// requests. If the handler recieves anything other than POST requests it will
// respond with a 405 Method Not Allowed.
func Handler(handler func(Report), cspHandler func(CSPReport)) safehttp.Handler {
	return safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		if r.Method() != safehttp.MethodPost {
			return w.WriteError(safehttp.StatusMethodNotAllowed)
		}

		b, err := ioutil.ReadAll(r.Body())
		if err != nil {
			return w.WriteError(safehttp.StatusBadRequest)
		}

		ct := r.Header.Get("Content-Type")
		if ct == "application/csp-report" {
			return handleDeprecatedCSPReports(cspHandler, w, b)
		} else if ct == "application/reports+json" {
			return handleReport(handler, w, b)
		}

		return w.WriteError(safehttp.StatusUnsupportedMediaType)
	})
}

func handleDeprecatedCSPReports(h func(CSPReport), w safehttp.ResponseWriter, b []byte) safehttp.Result {
	// In CSP2 it is clearly stated that a report has a single key 'csp-report'
	// which holds the report object. Like this:
	// {
	//   "csp-report": {
	//     // report goes here
	//   }
	// }
	// Source: https://www.w3.org/TR/CSP2/#violation-reports
	//
	// But in the CSP3 spec this 'csp-report' key is never mentioned. So the report
	// would look like this:
	// {
	//   // report goes here
	// }
	// Source: https://w3c.github.io/webappsec-csp/#deprecated-serialize-violation
	//
	// Because of this we have to support both. :/
	r := struct {
		CSPReport         json.RawMessage `json:"csp-report"`
		BlockedURL        string          `json:"blocked-uri"`
		Disposition       string          `json:"disposition"`
		DocumentURL       string          `json:"document-uri"`
		EffectiveDirectiv string          `json:"effective-directive"`
		OriginalPolicy    string          `json:"original-policy"`
		Referrer          string          `json:"referrer"`
		Sample            string          `json:"script-sample"`
		StatusCode        uint            `json:"status-code"`
		ViolatedDirective string          `json:"violated-directive"`
		SourceFile        string          `json:"source-file"`
		LineNo            uint            `json:"lineno"`
		LineNumber        uint            `json:"line-number"`
		ColNo             uint            `json:"colno"`
		ColumnNumber      uint            `json:"column-number"`
	}{}
	if err := json.Unmarshal(b, &r); err != nil {
		return w.WriteError(safehttp.StatusBadRequest)
	}

	if len(r.CSPReport) != 0 {
		if err := json.Unmarshal(r.CSPReport, &r); err != nil {
			return w.WriteError(safehttp.StatusBadRequest)
		}
	}

	ln := r.LineNo
	if ln == 0 {
		ln = r.LineNumber
	}
	cn := r.ColNo
	if cn == 0 {
		cn = r.ColumnNumber
	}

	rr := CSPReport{
		BlockedURL:         r.BlockedURL,
		Disposition:        r.Disposition,
		DocumentURL:        r.DocumentURL,
		EffectiveDirective: r.EffectiveDirectiv,
		OriginalPolicy:     r.OriginalPolicy,
		Referrer:           r.Referrer,
		Sample:             r.Sample,
		StatusCode:         r.StatusCode,
		ViolatedDirective:  r.ViolatedDirective,
		SourceFile:         r.SourceFile,
		LineNumber:         ln,
		ColumnNumber:       cn,
	}
	h(rr)

	return w.Write(safehttp.NoContentResponse{})
}

var reportHandlers = map[string]func(json.RawMessage) (body interface{}, ok bool){
	"csp-violation": cspViolationHandler,
}

func handleReport(h func(Report), w safehttp.ResponseWriter, b []byte) safehttp.Result {
	var rList []struct {
		Type      string          `json:"type"`
		Age       uint64          `json:"age"`
		URL       string          `json:"url"`
		UserAgent string          `json:"userAgent"`
		Body      json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal(b, &rList); err != nil {
		return w.WriteError(safehttp.StatusBadRequest)
	}

	badReport := false
	for _, r := range rList {
		reportToSend := Report{
			Type:      r.Type,
			Age:       r.Age,
			URL:       r.URL,
			UserAgent: r.UserAgent,
		}

		if f, ok := reportHandlers[r.Type]; ok {
			b, ok := f(r.Body)
			if !ok {
				badReport = true
				continue
			}
			reportToSend.Body = b
		} else {
			b := map[string]interface{}{}
			if err := json.Unmarshal(r.Body, &b); err != nil {
				badReport = true
				continue
			}
			reportToSend.Body = b
		}

		h(reportToSend)
	}

	if badReport {
		return w.WriteError(safehttp.StatusBadRequest)
	}

	return w.Write(safehttp.NoContentResponse{})
}

// cspViolationHandler parses reports of type csp-violation and returns a CSPReport.
func cspViolationHandler(m json.RawMessage) (body interface{}, ok bool) {
	// https://w3c.github.io/webappsec-csp/#reporting
	r := struct {
		BlockedURL         string `json:"blockedURL"`
		Disposition        string `json:"disposition"`
		DocumentURL        string `json:"documentURL"`
		EffectiveDirective string `json:"effectiveDirective"`
		OriginalPolicy     string `json:"originalPolicy"`
		Referrer           string `json:"referrer"`
		Sample             string `json:"sample"`
		StatusCode         uint   `json:"statusCode"`
		SourceFile         string `json:"sourceFile"`
		LineNumber         uint   `json:"lineNumber"`
		ColumnNumber       uint   `json:"columnNumber"`
	}{}
	if err := json.Unmarshal(m, &r); err != nil {
		return nil, false
	}

	return CSPReport{
		BlockedURL:         r.BlockedURL,
		Disposition:        r.Disposition,
		DocumentURL:        r.DocumentURL,
		EffectiveDirective: r.EffectiveDirective,
		OriginalPolicy:     r.OriginalPolicy,
		Referrer:           r.Referrer,
		Sample:             r.Sample,
		StatusCode:         r.StatusCode,
		// In CSP3 ViolatedDirective has been removed but is kept as
		// a copy of EffectiveDirective for backwards compatibility.
		ViolatedDirective: r.EffectiveDirective,
		SourceFile:        r.SourceFile,
		LineNumber:        r.LineNumber,
		ColumnNumber:      r.ColumnNumber,
	}, true
}
