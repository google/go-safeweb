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

var randReader = rand.Reader

// nonceSize is the size of the nonces in bytes.
const nonceSize = 20

func generateNonce() string {
	b := make([]byte, nonceSize)
	_, err := randReader.Read(b)
	if err != nil {
		// TODO: handle this better, what should happen here?
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// Policy defines a CSP policy.
type Policy struct {
	reportOnly bool

	// serialize serializes this policy for use in a Content-Security-Policy header
	// or in a Content-Security-Policy-Report-Only header. If needsNonce is true,
	// a nonce will be provided to serialize.
	serialize func(nonce string) string
}

type ctxKey struct{}

// Nonce retrieves the nonce from the given context. If there is no nonce stored
// in the context, an empty string is returned.
func Nonce(ctx context.Context) string {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return ""
	}
	return v.(string)
}

// StrictCSPBuilder can be used to build a strict, nonce-based CSP.
// See https://csp.withgoogle.com/docs/strict-csp.html for more info.
//
// If BaseURI is an empty string the base-uri directive will be set to 'none'.
type StrictCSPBuilder struct {
	ReportOnly    bool
	StrictDynamic bool
	UnsafeEval    bool
	BaseURI       string
	ReportURI     string
}

// Build creates a Policy based on the specified options.
func (s StrictCSPBuilder) Build() Policy {
	return Policy{
		reportOnly: s.ReportOnly,
		serialize: func(nonce string) string {
			var b strings.Builder

			b.WriteString("object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-")
			b.WriteString(nonce)
			b.WriteString("'")

			if s.StrictDynamic {
				b.WriteString(" 'strict-dynamic'")
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

// FramingPolicy creates a new CSP policy with frame-ancestors set to 'self'.
//
// TODO: allow relaxation on specific endpoints according to #77.
func FramingPolicy(reportOnly bool, reportURI string) Policy {
	return Policy{
		reportOnly: reportOnly,
		serialize: func(_ string) string {
			var b strings.Builder
			b.WriteString("frame-ancestors 'self'")

			if reportURI != "" {
				b.WriteString("; report-uri ")
				b.WriteString(reportURI)
			}

			return b.String()
		},
	}
}

// Interceptor intercepts requests and applies CSP policies.
type Interceptor struct {
	Policies []Policy
}

// NewInterceptor creates an interceptor from the provided policies.
func NewInterceptor(p ...Policy) Interceptor {
	return Interceptor{Policies: p}
}

// Default creates a new CSP interceptor with a strict nonce-based policy and a
// framing policy, both in enforcement mode.
func Default(reportURI string) Interceptor {
	return NewInterceptor(StrictCSPBuilder{ReportURI: reportURI}.Build(), FramingPolicy(false, reportURI))
}

// Before claims and sets the Content-Security-Policy header and the
// Content-Security-Policy-Report-Only header.
func (it Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	var csps []string
	var reportCsps []string
	nonce := generateNonce()
	r.SetContext(context.WithValue(r.Context(), ctxKey{}, nonce))
	for _, p := range it.Policies {
		v := p.serialize(nonce)
		if p.reportOnly {
			reportCsps = append(reportCsps, v)
		} else {
			csps = append(csps, v)
		}
	}

	h := w.Header()
	setCSP, err := h.Claim("Content-Security-Policy")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	setCSP(csps)

	setCSPReportOnly, err := h.Claim("Content-Security-Policy-Report-Only")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}
	setCSPReportOnly(reportCsps)

	return safehttp.Result{}
}
