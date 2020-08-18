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
const nonceSize = 8

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
	// or in a Content-Security-Policy-Report-Only header. If the given context
	// contains a nonce, it is used, otherwise a new one is generated and placed
	// in the context.
	serialize func(context.Context) (string, context.Context)
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

// NewStrictCSP creates a new strict, nonce-based CSP.
// See https://csp.withgoogle.com/docs/strict-csp.html for more info.
func NewStrictCSP(reportOnly bool, strictDynamic bool, unsafeEval bool, baseURI string, reportURI string) Policy {
	return Policy{
		reportOnly: reportOnly,
		serialize: func(ctx context.Context) (string, context.Context) {
			var b strings.Builder

			b.WriteString("object-src 'none'; script-src 'unsafe-inline' https: http: 'nonce-")
			n := Nonce(ctx)
			if n == "" {
				n = generateNonce()
				ctx = context.WithValue(ctx, ctxKey{}, n)
			}
			b.WriteString(n)
			b.WriteString("'")

			if strictDynamic {
				b.WriteString(" 'strict-dynamic'")
			}
			if unsafeEval {
				b.WriteString(" 'unsafe-eval'")
			}

			b.WriteString("; base-uri ")
			if baseURI == "" {
				b.WriteString("'none'")
			} else {
				b.WriteString(baseURI)
			}

			if reportURI != "" {
				b.WriteString("; report-uri ")
				b.WriteString(reportURI)
			}

			return b.String(), ctx
		},
	}
}

// NewFramingCSP creates a new CSP policy with frame-ancestors set to 'self'.
//
// TODO: allow relaxation on specific endpoints according to #77.
func NewFramingCSP(reportOnly bool) Policy {
	return Policy{
		reportOnly: reportOnly,
		serialize: func(ctx context.Context) (string, context.Context) {
			return "frame-ancestors 'self'", ctx
		},
	}
}

// Interceptor intercepts requests and applies CSP policies.
type Interceptor struct {
	Policies []Policy
}

// Default creates a new CSP interceptor with a strict nonce-based policy and a
// framing policy, both in enforcement mode.
func Default(reportURI string) Interceptor {
	return Interceptor{
		Policies: []Policy{
			NewStrictCSP(false, false, false, "", reportURI),
			NewFramingCSP(false),
		},
	}
}

// Before claims and sets the Content-Security-Policy header and the
// Content-Security-Policy-Report-Only header.
func (it Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	h := w.Header()
	setCSP, err := h.Claim("Content-Security-Policy")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}

	setCSPReportOnly, err := h.Claim("Content-Security-Policy-Report-Only")
	if err != nil {
		return w.ServerError(safehttp.StatusInternalServerError)
	}

	csps := make([]string, 0)
	reportCsps := make([]string, 0)
	for _, p := range it.Policies {
		v, ctx := p.serialize(r.Context())
		r.SetContext(ctx)
		if p.reportOnly {
			reportCsps = append(reportCsps, v)
		} else {
			csps = append(csps, v)
		}
	}
	setCSP(csps)
	setCSPReportOnly(reportCsps)

	return safehttp.Result{}
}
