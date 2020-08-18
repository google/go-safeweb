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
	dir       Directive
	vals      []string
	nonces    bool
	lastNonce string
	// readRand is used for dependency injection in tests.
	readRand func([]byte) (int, error)
}

// NewPolicyDirective creates a new PolicyDirective given the directive, the
// values and whether or not a nonce directive should be added.
func NewPolicyDirective(d Directive, values []string, nonces bool) *PolicyDirective {
	return &PolicyDirective{
		dir:      d,
		vals:     values,
		nonces:   nonces,
		readRand: rand.Read,
	}
}

// nonceSize is the size of the nonces in bytes.
const nonceSize = 8

func (pd *PolicyDirective) generateNonce() {
	b := make([]byte, nonceSize)
	_, err := pd.readRand(b)
	if err != nil {
		// TODO: handle this better, what should happen here?
		panic(err)
	}
	pd.lastNonce = base64.StdEncoding.EncodeToString(b)
}

// Policy defines a CSP policy, containing many directives.
type Policy struct {
	Directives []*PolicyDirective
}

// NewPolicy creates a new strict, nonce-based CSP.
// See https://csp.withgoogle.com/docs/strict-csp.html for more info.
//
// TODO: maybe reportURI should be safehttp.URL?
func NewPolicy(reportURI string) *Policy {
	return &Policy{
		Directives: []*PolicyDirective{
			NewPolicyDirective(DirectiveObjectSrc, []string{ValueNone}, false),
			NewPolicyDirective(DirectiveScriptSrc, []string{
				ValueUnsafeInline,
				ValueUnsafeEval,
				ValueStrictDynamic,
				ValueHTTPS,
				ValueHTTP,
			}, true),
			NewPolicyDirective(DirectiveBaseURI, []string{ValueNone}, false),
			NewPolicyDirective(DirectiveReportURI, []string{reportURI}, false),
		},
	}
}

// Serialize serializes this policy for use in a Content-Security-Policy header
// or in a Content-Security-Policy-Report-Only header. The nonces generated for
// each directive is also returned.
func (p Policy) Serialize() (csp string, nonces map[Directive]string) {
	nonces = make(map[Directive]string)
	b := strings.Builder{}

	for i, d := range p.Directives {
		if i != 0 {
			b.WriteString("; ")
		}
		b.WriteString(string(d.dir))

		if d.nonces {
			d.generateNonce()
			b.WriteString(" 'nonce-")
			b.WriteString(d.lastNonce)
			b.WriteString("'")
			nonces[d.dir] = d.lastNonce
		}

		for _, v := range d.vals {
			b.WriteString(" ")
			b.WriteString(v)
		}
	}
	return b.String(), nonces
}

// Interceptor intercepts requests and applies CSP policies.
type Interceptor struct {
	// EnforcementPolicy will be applied as the Content-Security-Policy header.
	EnforcementPolicy *Policy

	// ReportOnlyPolicy will be applied as the
	// Content-Security-Policy-Report-Only header.
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
