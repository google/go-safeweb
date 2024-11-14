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

// Package uncheckedconversions provides functions to create values of package safesql types from plain strings.
// Uses of these functions could potentially result in instances of safesql types that violate their type contracts, and hence result in security vulnerabilities.
package uncheckedconversions

import (
	"github.com/google/go-safeweb/safesql"
	"github.com/google/go-safeweb/safesql/internal/raw"
)

var trustedSQLStringCtor = raw.TrustedSQLString.(func(string) safesql.TrustedSQLString)

// TrustedSQLStringFromStringKnownToSatisfyTypeContract promotes the given string to a trusted string.
// Only strings known to be under the programmer control and trusted strings should be passed to this function.
//
// One example of correct use of this function would be to cast a query that was retrieved from a query storage to be used with the safesql package.
// If the query storage is under the programmer control and user input cannot be put into it then the string is known to satisfy the type contract.
func TrustedSQLStringFromStringKnownToSatisfyTypeContract(trusted string) safesql.TrustedSQLString {
	return trustedSQLStringCtor(trusted)
}
