// Copyright 2022 Google LLC
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

// Package unsafecspfortests can be used to disable CSP on specific handler registration in tests.
//
// This package should only be used in tests.
package unsafecspfortests

import "github.com/google/go-safeweb/safehttp/plugins/csp/internalunsafecsp"

type endlessAReader struct{}

func (endlessAReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 41
	}
	return len(b), nil
}

// UseStaticRandom will make all nonces assume a static value.
func UseStaticRandom() {
	internalunsafecsp.RandReader = endlessAReader{}
}

// DisableStrict completely disables Strict CSP.
func DisableStrict() internalunsafecsp.DisableStrict {
	return internalunsafecsp.DisableStrict{SkipReports: true}
}

// DisableTrustedTypes completely disables TrustedTypes
func DisableTrustedTypes() internalunsafecsp.DisableTrustedTypes {
	return internalunsafecsp.DisableTrustedTypes{SkipReports: true}
}
