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

// Package raw is used to provide a bypass mechanism to implement unchecked and legacy conversions packages.
// This package works as a proxy between safesql and any other "conversions" package.
//
// The way it functions is to expect safesql to provide the unexported constructors for TrustedSQLString at init() time.
// Since this package is in internal/ it can only be imported by a parent package, so it is known at compile time that
// these constructors are not unsafely passed around.
package raw

// TrustedSQLString is the constructor for a TrustedSQLString to be used by the unchecked and legacy conversions packages.
// This variable will be assigned by the safesql package at init time.
// The reason why this is an empty interface is to avoid cyclic dependency between safeslq and this package.
var TrustedSQLString interface{}
