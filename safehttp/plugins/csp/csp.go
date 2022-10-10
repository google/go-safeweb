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

// Package csp provides a safehttp.Interceptor which applies Content-Security Policies
// to responses.
//
// These default policies are provided:
//   - A strict nonce based CSP
//   - A framing policy which sets frame-ancestors to 'self'
//   - A Trusted Types policy which makes usage of dangerous web API functions secure by default
package csp

// TODO(empijei): add support for report-to and report groups.

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp"
	"github.com/google/go-safeweb/safehttp/plugins/htmlinject"

	"github.com/google/go-safeweb/safehttp"
)

const (
	responseHeaderKey           = "Content-Security-Policy"
	responseHeaderReportOnlyKey = responseHeaderKey + "-Report-Only"
)

// nonceSize is the size of the nonces in bytes. According to the CSP3 spec it should
// be larger than 16 bytes. 20 bytes was picked to be future proof.
// https://www.w3.org/TR/CSP3/#security-nonces
const nonceSize = 20

func generateNonce() string {
	b := make([]byte, nonceSize)
	_, err := internalunsafecsp.RandReader.Read(b)
	if err != nil {
		panic(fmt.Errorf("failed to generate entropy using crypto/rand/RandReader: %v", err))
	}
	return base64.StdEncoding.EncodeToString(b)
}

type key string

const (
	nonceKey   key = "csp-nonce"
	headersKey key = "csp-headers"
)

// Nonce retrieves the nonce from the given context. If there is no nonce stored
// in the context, an error will be returned.
func Nonce(ctx context.Context) (string, error) {
	v := safehttp.FlightValues(ctx).Get(nonceKey)
	if v == nil {
		return "", errors.New("no nonce in context")
	}
	return v.(string), nil
}

// nonce retrieves the nonces from the request.
// If none is available, one will be generated and added to it.
func nonce(r *safehttp.IncomingRequest) string {
	v := safehttp.FlightValues(r.Context()).Get(nonceKey)
	var nonce string
	if v == nil {
		nonce = generateNonce()
		safehttp.FlightValues(r.Context()).Put(nonceKey, nonce)
	} else {
		nonce = v.(string)
	}
	return nonce
}

func claimedHeaders(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) (cspe func([]string), cspro func([]string)) {
	type claimed struct {
		cspe, cspro func([]string)
	}
	v := safehttp.FlightValues(r.Context()).Get(headersKey)
	var c claimed
	if v == nil {
		h := w.Header()
		cspe := h.Claim(responseHeaderKey)
		cspro := h.Claim(responseHeaderReportOnlyKey)
		c = claimed{cspe: cspe, cspro: cspro}
		safehttp.FlightValues(r.Context()).Put(headersKey, c)
	} else {
		c = v.(claimed)
	}
	return c.cspe, c.cspro
}

// Policy defines a CSP policy.
type Policy interface {
	// Serialize serializes this policy for use in a Content-Security-Policy header
	// or in a Content-Security-Policy-Report-Only header. A nonce will be provided
	// to Serialize which can be used in 'nonce-{random-nonce}' values in directives.
	// If a config has matched the interceptor, it will also be passed.
	Serialize(nonce string, cfg safehttp.InterceptorConfig) string
	// Match allows to match configurations that are specific to this policy.
	Match(cfg safehttp.InterceptorConfig) bool
	// Overridden is used to check if a configuration is overriding the policy.
	Overridden(cfg safehttp.InterceptorConfig) (disabled, reportOnly bool)
}

func report(reportURI string) string {
	var b strings.Builder

	if reportURI != "" {
		b.WriteString("report-uri ")
		b.WriteString(reportURI)
		b.WriteString("; ")
	}

	return b.String()
}

// Interceptor intercepts requests and applies CSP policies.
// Multiple interceptors can be installed at the same time.
type Interceptor struct {
	// Policy is the policy the interceptor should enforce.
	Policy Policy
	// ReportOnly makes Policy be set report-only.
	ReportOnly bool
}

var _ safehttp.Interceptor = Interceptor{}

// Default creates new CSP interceptors with a strict nonce-based policy and a TrustedTypes policy,
// all in enforcement mode.
// Framing policies are installed by the framing interceptor.
func Default(reportURI string) []Interceptor {
	return []Interceptor{
		{Policy: StrictPolicy{ReportURI: reportURI}},
		{Policy: TrustedTypesPolicy{ReportURI: reportURI}},
	}
}

func (it Interceptor) processOverride(cfg safehttp.InterceptorConfig, nonce string) (enf, ro string) {
	disabled, reportOnly := false, false
	if it.Policy.Match(cfg) {
		disabled, reportOnly = it.Policy.Overridden(cfg)
	}
	if disabled {
		return "", ""
	}
	p := it.Policy.Serialize(nonce, cfg)
	if reportOnly || it.ReportOnly {
		return "", p
	}
	return p, ""
}

// Before claims and sets the Content-Security-Policy header and the
// Content-Security-Policy-Report-Only header.
func (it Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	nonce := nonce(r)
	enf, ro := it.processOverride(cfg, nonce)
	setCSP, setCSPReportOnly := claimedHeaders(w, r)
	if enf != "" {
		prev := w.Header().Values(responseHeaderKey)
		setCSP(append(prev, enf))
	}
	if ro != "" {
		prev := w.Header().Values(responseHeaderReportOnlyKey)
		setCSPReportOnly(append(prev, ro))
	}
	return safehttp.NotWritten()
}

// Commit adds the nonce to the safehttp.TemplateResponse which is going to be
// injected as the value of the nonce attribute in <script> and <link> tags. The
// nonce is going to be unique for each safehttp.IncomingRequest.
func (it Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) {
	tmplResp, ok := resp.(*safehttp.TemplateResponse)
	if !ok {
		return
	}

	nonce, err := Nonce(r.Context())
	if err != nil {
		// The nonce should have been added in the Before stage and, if that is
		// not the case, a server misconfiguration occurred.
		panic("no CSP nonce")
	}

	if tmplResp.FuncMap == nil {
		tmplResp.FuncMap = map[string]interface{}{}
	}
	tmplResp.FuncMap[htmlinject.CSPNoncesDefaultFuncName] = func() string { return nonce }
}

// Match returns false since there are no supported configurations.
func (it Interceptor) Match(cfg safehttp.InterceptorConfig) bool {
	return it.Policy.Match(cfg)
}
