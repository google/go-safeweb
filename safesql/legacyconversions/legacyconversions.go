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

// Package legacyconversions provides functions to create values of package safesql types from plain strings.
// Use of these functions could potentially result in instances of safesql types that violate their type contracts, and hence result in security vulnerabilities.
// This package should only be used to gradually migrate to use the safesql package but every use of it should eventually be removed as it represents a security risk.
package legacyconversions

import (
	"github.com/google/go-safeweb/safesql"
	"github.com/google/go-safeweb/safesql/internal/raw"
)

var trustedSQLStringCtor = raw.TrustedSQLString.(func(string) safesql.TrustedSQLString)

// RiskilyAssumeTrustedSQLString riskily promotes the given string to a trusted string.
// Uses of this function should only be used when migrating to use safesql and should eventually be removed.
func RiskilyAssumeTrustedSQLString(trusted string) safesql.TrustedSQLString {
	return trustedSQLStringCtor(trusted)
}
