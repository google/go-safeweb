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

package csp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/go-safeweb/safehttp"
)

// Directive is the name of a single CSP directive.
type Directive string

const (
	DirectiveScriptSrc Directive = "script-src"
	DirectiveStyleSrc            = "style-src"
	DirectiveObjectSrc           = "object-src"
	DirectiveBaseURI             = "base-uri"
	DirectiveReportURI           = "report-uri"
)

const (
	ValueHTTPS         = "https:"
	ValueHTTP          = "http:"
	ValueUnsafeEval    = "'unsafe-eval'"
	ValueUnsafeInline  = "'unsafe-inline'"
	ValueNone          = "'none'"
	ValueStrictDynamic = "'strict-dynamic'"
)

// PolicyDirective contains a single CSP directive.
type PolicyDirective struct {
	Directive Directive
	Values    []string
	AddNonce  bool
}

// nonceSize is the size of the nonces in bytes.
const nonceSize = 8

func generateNonce(readRand func([]byte) (int, error)) string {
	if readRand == nil {
		readRand = rand.Read
	}
	b := make([]byte, nonceSize)
	_, err := readRand(b)
	if err != nil {
		// TODO: handle this better, what should happen here?
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// Policy defines a CSP policy, containing many directives.
type Policy struct {
	Directives []*PolicyDirective
	// readRand is used for dependency injection in tests.
	readRand func([]byte) (int, error)
}

// NewPolicy creates a new strict, nonce-based CSP.
// See https://csp.withgoogle.com/docs/strict-csp.html for more info.
//
// TODO: maybe reportURI should be safehttp.URL?
func NewPolicy(reportURI string) *Policy {
	return &Policy{
		Directives: []*PolicyDirective{
			{Directive: DirectiveObjectSrc, Values: []string{ValueNone}, AddNonce: false},
			{
				Directive: DirectiveScriptSrc,
				Values: []string{
					ValueUnsafeInline,
					ValueUnsafeEval,
					ValueStrictDynamic,
					ValueHTTPS,
					ValueHTTP,
				},
				AddNonce: true,
			},
			{Directive: DirectiveBaseURI, Values: []string{ValueNone}, AddNonce: false},
			{Directive: DirectiveReportURI, Values: []string{reportURI}, AddNonce: false},
		},
	}
}

// Serialize serializes this policy for use in a Content-Security-Policy header
// or in a Content-Security-Policy-Report-Only header. The nonces generated for
// each directive are also returned.
func (p Policy) Serialize() (csp string, nonces map[Directive]string) {
	nonces = make(map[Directive]string)
	values := make([]string, 0, len(p.Directives))

	for _, d := range p.Directives {
		var b strings.Builder
		b.WriteString(string(d.Directive))

		if d.AddNonce {
			n := generateNonce(p.readRand)
			b.WriteString(fmt.Sprintf(" 'nonce-%s'", n))
			nonces[d.Directive] = n
		}

		b.WriteString(" ")
		b.WriteString(strings.Join(d.Values, " "))

		values = append(values, b.String())
	}

	return strings.Join(values, "; "), nonces
}

// Interceptor intercepts requests and applies CSP policies.
type Interceptor struct {
	// EnforcementPolicy will be applied as the Content-Security-Policy header.
	EnforcementPolicy *Policy

	// ReportOnlyPolicy will be applied as the Content-Security-Policy-Report-Only
	// header.
	ReportOnlyPolicy *Policy
}

// Default creates a new CSP interceptor with a strict nonce-based policy in
// enforcement mode.
func Default(reportURI string) Interceptor {
	return Interceptor{EnforcementPolicy: NewPolicy(reportURI)}
}

type ctxKey string

// Before claims and sets the Content-Security-Policy header and the
// Content-Security-Policy-Report-Only header.
func (it Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	h := w.Header()
	setCSP, err := h.Claim("Content-Security-Policy")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	if it.EnforcementPolicy != nil {
		s, nonces := it.EnforcementPolicy.Serialize()
		setCSP([]string{s})
		if len(nonces) != 0 {
			r.SetContext(context.WithValue(r.Context(), ctxKey("enforce"), nonces))
		}
	}

	setCSPReportOnly, err := h.Claim("Content-Security-Policy-Report-Only")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	if it.ReportOnlyPolicy != nil {
		s, nonces := it.ReportOnlyPolicy.Serialize()
		setCSPReportOnly([]string{s})
		if len(nonces) != 0 {
			r.SetContext(context.WithValue(r.Context(), ctxKey("report"), nonces))
		}
	}

	return safehttp.Result{}
}
