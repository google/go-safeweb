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
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-safeweb/safehttp"
)

var randReader = rand.Reader

// nonceSize is the size of the nonces in bytes. According to the CSP3 spec it should
// be larger than 16 bytes. 20 bytes was picked to be future proof.
// https://www.w3.org/TR/CSP3/#security-nonces
const nonceSize = 20

func generateNonce() string {
	b := make([]byte, nonceSize)
	_, err := randReader.Read(b)
	if err != nil {
		panic(fmt.Errorf("failed to generate entropy using crypto/rand/RandReader: %v", err))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// Policy defines a CSP policy.
type Policy struct {
	// serialize serializes this policy for use in a Content-Security-Policy header
	// or in a Content-Security-Policy-Report-Only header. A nonce will be provided
	// to serialize which can be used in 'nonce-{random-nonce}' values in directives.
	serialize func(nonce string) string
}

type ctxKey struct{}

// Nonce retrieves the nonce from the given context. If there is no nonce stored
// in the context, an error will be returned.
func Nonce(ctx context.Context) (string, error) {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return "", errors.New("no nonce in context")
	}
	return v.(string), nil
}

// StrictCSPBuilder can be used to build a strict, nonce-based CSP.
//
// See https://csp.withgoogle.com/docs/strict-csp.html for more info.
type StrictCSPBuilder struct {
	// NoStrictDynamic controls whether script-src should contain the 'strict-dynamic'
	// value.
	//
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/script-src#strict-dynamic
	// for more info.
	NoStrictDynamic bool
	// UnsafeEval controls whether script-src should contain the 'unsafe-eval' value.
	// If enabled, the eval() JavaScript function is allowed.
	UnsafeEval bool
	// BaseURI controls the base-uri directive. If BaseURI is an empty string the
	// directive will be set to 'none'. The base-uri directive restricts the URLs
	// which can be used in a document's <base> element.
	//
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/base-uri
	// for more info.
	BaseURI string
	// ReportURI controls the report-uri directive. If ReportUri is empty, no report-uri
	// directive will be set.
	ReportURI string
}

// Build creates a Policy based on the specified options.
func (s StrictCSPBuilder) Build() Policy {
	return Policy{
		serialize: func(nonce string) string {
			var b strings.Builder

			// object-src 'none'; script-src 'unsafe-inline' 'nonce-{random}'
			b.WriteString("object-src 'none'; script-src 'unsafe-inline' 'nonce-")
			b.WriteString(nonce)
			b.WriteByte('\'')

			if !s.NoStrictDynamic {
				b.WriteString(" 'strict-dynamic' https: http:")
			}

			if s.UnsafeEval {
				b.WriteString(" 'unsafe-eval'")
			}

			b.WriteString("; base-uri ")
			if s.BaseURI == "" {
				b.WriteString("'none'")
			} else {
				b.WriteString(s.BaseURI)
			}

			if s.ReportURI != "" {
				b.WriteString("; report-uri ")
				b.WriteString(s.ReportURI)
			}

			return b.String()
		},
	}
}

// FramingPolicyBuilder can be used to create a new CSP policy with frame-ancestors
// set to 'self'.
//
// TODO: allow relaxation on specific endpoints according to #77.
type FramingPolicyBuilder struct {
	// ReportURI controls the report-uri directive. If ReportUri is empty, no report-uri
	// directive will be set.
	ReportURI string
}

// Build creates a Policy based on the specified options.
func (f FramingPolicyBuilder) Build() Policy {
	return Policy{
		serialize: func(_ string) string {
			var b strings.Builder
			b.WriteString("frame-ancestors 'self'")

			if f.ReportURI != "" {
				b.WriteString("; report-uri ")
				b.WriteString(f.ReportURI)
			}

			return b.String()
		},
	}
}

// Interceptor intercepts requests and applies CSP policies.
type Interceptor struct {
	// Enforce specifies which policies will be set as the Content-Security-Policy
	// header.
	Enforce []Policy
	// ReportOnly specifies which policies will be set as the Content-Security-Policy-Report-Only
	// header.
	ReportOnly []Policy
}

// Default creates a new CSP interceptor with a strict nonce-based policy and a
// framing policy, both in enforcement mode.
func Default(reportURI string) Interceptor {
	return Interceptor{
		Enforce: []Policy{
			StrictCSPBuilder{ReportURI: reportURI}.Build(),
			FramingPolicyBuilder{ReportURI: reportURI}.Build(),
		},
	}
}

// Before claims and sets the Content-Security-Policy header and the
// Content-Security-Policy-Report-Only header.
func (it Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	nonce := generateNonce()
	r.SetContext(context.WithValue(r.Context(), ctxKey{}, nonce))

	var CSPs []string
	for _, p := range it.Enforce {
		CSPs = append(CSPs, p.serialize(nonce))
	}
	var reportCSPs []string
	for _, p := range it.ReportOnly {
		reportCSPs = append(reportCSPs, p.serialize(nonce))
	}

	h := w.Header()
	setCSP, err := h.Claim("Content-Security-Policy")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	setCSPReportOnly, err := h.Claim("Content-Security-Policy-Report-Only")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}

	setCSP(CSPs)
	setCSPReportOnly(reportCSPs)

	return safehttp.Result{}
}
