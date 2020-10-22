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

// Package safesql implements a safe version of the standard sql package while trying to keep the API as similar as
// possible to the original one.
// The concept of this package is to provide "safe by construction" SQL strings so that code that would accidentally introduce
// SQL injection vulnerabilities does not compile.
// If uncheckedconversions and legacyconversions are not used and the sql package is forbidden this package guarantees that only compile-time constants will be
// interpreted as SQL, thus preventing attacker-controlled strings to be accidentally executed.
//
// Migration Examples
//
// Code like the following is trivial to migrate from sql to safesql:
// 	db.Query("SELECT ...", args...)
// The only change required would be to promote the string literal to a trusted string:
// 	db.Query(safesql.New("SELECT ..."), args...)
// For more complicated cases it might be needed to use the helper functions like Join and Concat.
// If the queries for the service are stored in a trusted runtime-only source that cannot be controlled by a user
// the uncheckedconversions package can be used to assert that those strings are under the programmer control.
// Note that unchecked conversions should be very limited, ideally never used, as they pose a security risk.
//
// Note on API documentation.
//
// For documentation on methods and types that wrap the standard ones please refer to the stdlib package doc instead, as
// all the types exported by this package are tiny wrappers around the standard ones and thus follow their behavior.
// The only relevant difference is that functions accept TrustedSQLString instances instead of plain "strings" and that some
// dangerous methods have been removed.
//
// Explainer
//
// This package wraps the sql package and all methods that would normally take a string take a TrustedSQLString instead.
// The constructor for TrustedSQLString takes a stringConstant as an argument, which is an unexported type constituted by a named string.
// The only way for a package outside of safesql to construct a TrustedSQLString is thus to pass an untyped string (only const strings can be untyped) to the constructor.
package safesql

import (
	"strconv"
	"strings"

	"github.com/google/go-safeweb/safesql/internal/raw"
)

func init() {
	// Initialize the bypass mechanisms for unchecked and legacy conversions.
	raw.TrustedSQLString = func(unsafe string) TrustedSQLString { return TrustedSQLString{unsafe} }
}

type stringConstant string

// TrustedSQLString is a string representing a SQL query that is known to be safe and not contain potentially malicious inputs.
type TrustedSQLString struct {
	s string
}

// New constructs a TrustedSQLString from a compile-time constant string.
// Since the stringConstant type is unexported the only way to call this function outside of this package is to pass
// a string literal or an untyped string const.
func New(text stringConstant) TrustedSQLString { return TrustedSQLString{string(text)} }

// NewFromUint64 constructs a TrustedSQLString from a uint64.
func NewFromUint64(i uint64) TrustedSQLString { return TrustedSQLString{strconv.FormatUint(i, 10)} }

// TrustedSQLStringConcat concatenates the given trusted SQL strings into a trusted string.
//
// Note: this function should not be abused to create arbitrary queries from user input, it is just
// intended as a helper to compose queries at runtime to avoid redundant constants.
func TrustedSQLStringConcat(ss ...TrustedSQLString) TrustedSQLString {
	return TrustedSQLStringJoin(ss, TrustedSQLString{})
}

// TrustedSQLStringJoin joins the given trusted SQL with the given separator the same way strings.Join would.
//
// Note: this function should not be abused to create arbitrary queries from user input, it is just
// intended as a helper to compose queries at runtime to avoid redundant constants.
func TrustedSQLStringJoin(ss []TrustedSQLString, sep TrustedSQLString) TrustedSQLString {
	accum := make([]string, 0, len(ss))
	for _, s := range ss {
		accum = append(accum, s.s)
	}
	return TrustedSQLString{strings.Join(accum, sep.s)}
}
func (t TrustedSQLString) String() string {
	return t.s
}

// TrustedSQLStringSplit functions as strings.Split but for TrustedSQLStrings.
func TrustedSQLStringSplit(s TrustedSQLString, sep TrustedSQLString) []TrustedSQLString {
	spl := strings.Split(s.s, sep.s)
	accum := make([]TrustedSQLString, 0, len(spl))
	for _, s := range spl {
		accum = append(accum, TrustedSQLString{s})
	}
	return accum
}
