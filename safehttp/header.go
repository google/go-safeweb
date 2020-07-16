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

package safehttp

import (
	"errors"
	"net/http"
	"net/textproto"
)

var disallowedHeaders = map[string]bool{"Set-Cookie": true}

const (
	disallowedErrorMessage = "disallowed header"
	immutableErrorMessage  = "immutable header"
)

// Header represents the key-value pairs in an HTTP header.
// The keys will be in canonical form, as returned by
// textproto.CanonicalMIMEHeaderKey.
type Header struct {
	wrapped   http.Header
	immutable map[string]bool
}

func newHeader(h http.Header) Header {
	return Header{wrapped: h, immutable: map[string]bool{}}
}

// MarkImmutable marks the header with the given name as immutable.
// This header is now read-only. The name is first canonicalized
// using textproto.CanonicalMIMEHeaderKey.
func (h Header) MarkImmutable(name string) {
	name = textproto.CanonicalMIMEHeaderKey(name)
	h.immutable[name] = true
}

// Set sets the header with the given name to the given value.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// If this headers is not immutable, this function removes all other
// values currently associated with this header before setting the new
// value. Returns an error when applied on immutable headers.
func (h Header) Set(name, value string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return errors.New(disallowedErrorMessage)
	}
	if h.immutable[name] {
		return errors.New(immutableErrorMessage)
	}
	h.wrapped.Set(name, value)
	return nil
}

// Add adds a new header with the given name and the given value to
// the collection of headers. The name is first canonicalized using
// textproto.CanonicalMIMEHeaderKey. Returns an error when applied
// on immutable headers.
func (h Header) Add(name, value string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return errors.New(disallowedErrorMessage)
	}
	if h.immutable[name] {
		return errors.New(immutableErrorMessage)
	}
	h.wrapped.Add(name, value)
	return nil
}

// Del deletes all headers with the given name. The name is first
// canonicalized using textproto.CanonicalMIMEHeaderKey. Returns an
// error when applied on immutable headers.
func (h Header) Del(name string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return errors.New(disallowedErrorMessage)
	}
	if h.immutable[name] {
		return errors.New(immutableErrorMessage)
	}
	h.wrapped.Del(name)
	return nil
}

// Get returns the value of the first header with the given name.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// If no header exists with the given name then "" is returned.
func (h Header) Get(name string) string {
	return h.wrapped.Get(name)
}

// Values returns all the values of all the headers with the given name.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// If no header exists with the name `name` then nil is returned.
func (h Header) Values(name string) []string {
	return h.wrapped.Values(name)
}

// SetCookie adds the cookie provided as a Set-Cookie header in the header
// collection.
// TODO: Replace http.Cookie with safehttp.Cookie.
func (h Header) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		h.wrapped.Add("Set-Cookie", v)
	}
}

// TODO: Add Write, WriteSubset and Clone when needed.
