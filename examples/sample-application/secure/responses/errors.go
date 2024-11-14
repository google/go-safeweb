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

//go:build go1.16
// +build go1.16

package responses

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

// Error is a safe error response (as recognized by the secure.dispatcher).
//
// This showcases implementing custom safe responses. See
// https://pkg.go.dev/github.com/google/safehtml/template#hdr-Threat_model.
type Error struct {
	StatusCode safehttp.StatusCode
	Message    safehtml.HTML
}

// NewError creates a new error response.
func NewError(code safehttp.StatusCode, message safehtml.HTML) Error {
	return Error{
		StatusCode: code,
		Message:    message,
	}
}

// Code returns the HTTP response code.
func (e Error) Code() safehttp.StatusCode {
	return e.StatusCode
}
