// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package internalunsafecsp is used internally to override CSP.
package internalunsafecsp

import "crypto/rand"

// RandReader is a source of random.
var RandReader = rand.Reader

// DisableTrustedTypes switches TT to report-only.
type DisableTrustedTypes struct {
	// SkipReports completely disables TT.
	SkipReports bool
}

// DisableStrict switches Strict CSP to report-only.
type DisableStrict struct {
	// SkipReports completely disables Strict CSP.
	SkipReports bool
}

// Framing is overridden by the framing override.
