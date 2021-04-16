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
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSet(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("Foo-Key", "Bar-Value")
	if got, want := h.Get("Foo-Key"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("Foo-Key") got: %q want %q`, got, want)
	}
}

// TestSetCanonicalization verifies that names of headers
// are canonicalized before being interpreted as header
// names.
// Note that the casing of the header name is different
// when accessing and modifying the same header.
func TestSetCanonicalization(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("fOo-KeY", "Bar-Value")
	if got, want := h.Get("FoO-kEy"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("FoO-kEy") got: %q want %q`, got, want)
	}
}

func TestSetEmptySetCookie(t *testing.T) {
	h := NewHeader(http.Header{})
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Set("Set-Cookie", "x=y") expected panic`)
	}()
	h.Set("Set-Cookie", "x=y")
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestSetClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("Foo-Key", "Pizza-Value")
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Set("Foo-Key", "Bar-Value") expected panic`)
	}()
	h.Set("Foo-Key", "Bar-Value")
	if diff := cmp.Diff([]string{"Pizza-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestSetEmptyClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Set("Foo-Key", "Bar-Value") expected panic`)
	}()
	h.Set("Foo-Key", "Bar-Value")
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestAdd(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Add("Foo-Key", "Bar-Value")
	h.Add("Foo-Key", "Pizza-Value")
	if diff := cmp.Diff([]string{"Bar-Value", "Pizza-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

// TestAddCanonicalization verifies that names of headers
// are canonicalized before being interpreted as header
// names.
// Note that the casing of the header name is different
// when accessing and modifying the same header.
func TestAddCanonicalization(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Add("fOo-KeY", "Bar-Value")
	h.Add("FoO-kEy", "Pizza-Value")
	if diff := cmp.Diff([]string{"Bar-Value", "Pizza-Value"}, h.Values("fOO-KEY")); diff != "" {
		t.Errorf("h.Values(\"fOO-KEY\")) mismatch (-want +got):\n%s", diff)
	}
}

func TestAddEmptySetCookie(t *testing.T) {
	h := NewHeader(http.Header{})
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Add("Set-Cookie", "x=y") expected panic`)
	}()
	h.Add("Set-Cookie", "x=y")
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestAddClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Add("Foo-Key", "Bar-Value")
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Add("Foo-Key", "Pizza-Value") expected panic`)
	}()
	h.Add("Foo-Key", "Pizza-Value")
	if diff := cmp.Diff([]string{"Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestAddEmptyClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Add("Foo-Key", "Pizza-Value") expected panic`)
	}()
	h.Add("Foo-Key", "Pizza-Value")
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDel(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("Foo-Key", "Bar-Value")
	h.Del("Foo-Key")
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

// TestDelCanonicalization verifies that names of headers
// are canonicalized before being interpreted as header
// names.
// Note that the casing of the header name is different
// when accessing and modifying the same header.
func TestDelCanonicalization(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("fOo-KeY", "Bar-Value")
	h.Del("FoO-kEy")
	if diff := cmp.Diff([]string{}, h.Values("FOO-kEY")); diff != "" {
		t.Errorf("h.Values(\"FOO-kEY\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelEmptySetCookie(t *testing.T) {
	h := NewHeader(http.Header{})
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Del("Set-Cookie") expected panic`)
	}()
	h.Del("Set-Cookie")
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("Foo-Key", "Bar-Value")
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Del("Foo-Key") expected panic`)
	}()
	h.Del("Foo-Key")
	if diff := cmp.Diff([]string{"Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelEmptyClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Del("Foo-Key") expected panic`)
	}()
	h.Del("Foo-Key")
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

// TestValuesModifyClaimed verifies that modifying the
// slice returned by Values() doesn't modify the underlying
// slice. The test ensures that Values() returns a copy
// of the underlying slice.
func TestValuesModifyClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Set("Foo-Key", "Bar-Value")
	h.Claim("Foo-Key")
	v := h.Values("Foo-Key")
	if diff := cmp.Diff([]string{"Bar-Value"}, v); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
	v[0] = "Evil-Value"
	if got, want := h.Get("Foo-Key"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("Foo-Key") got: %v want: %v`, got, want)
	}
}

// TestValuesOrdering ensures that the Values() function
// return the headers values in the order that they were
// set.
func TestValuesOrdering(t *testing.T) {
	var tests = []struct {
		name   string
		values []string
	}{
		{
			name:   "Bar Pizza",
			values: []string{"Bar-Value", "Pizza-Value"},
		},
		{
			name:   "Pizza Bar",
			values: []string{"Pizza-Value", "Bar-Value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeader(http.Header{})
			h.Add("Foo-Key", tt.values[0])
			h.Add("Foo-Key", tt.values[1])
			if diff := cmp.Diff(tt.values, h.Values("Foo-Key")); diff != "" {
				t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestManyEqualKeyValuePairs(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Add("Foo-Key", "Bar-Value")
	h.Add("Foo-Key", "Bar-Value")
	if diff := cmp.Diff([]string{"Bar-Value", "Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

// TestAddSet tests that calling Set() after calling Add() will
// correctly remove the previously added header before setting
// the new one.
func TestAddSet(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Add("Foo-Key", "Bar-Value")
	h.Set("Foo-Key", "Pizza-Value")
	if diff := cmp.Diff([]string{"Pizza-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestValuesEmptyHeader(t *testing.T) {
	h := NewHeader(http.Header{})
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestClaim(t *testing.T) {
	h := NewHeader(http.Header{})
	set := h.Claim("Foo-Key")
	set([]string{"Bar-Value", "Pizza-Value"})
	if diff := cmp.Diff([]string{"Bar-Value", "Pizza-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

// TestClaimCanonicalization verifies that names of headers
// are canonicalized before being interpreted as header
// names.
// Note that the casing of the header name is different
// when accessing and modifying the same header.
func TestClaimCanonicalization(t *testing.T) {
	h := NewHeader(http.Header{})
	set := h.Claim("fOO-kEY")
	set([]string{"Bar-Value", "Pizza-Value"})
	if diff := cmp.Diff([]string{"Bar-Value", "Pizza-Value"}, h.Values("fOo-kEy")); diff != "" {
		t.Errorf("h.Values(\"fOo-kEy\") mismatch (-want +got):\n%s", diff)
	}
}

func TestClaimSetCookie(t *testing.T) {
	h := NewHeader(http.Header{})
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Claim("Set-Cookie") expected panic`)
	}()
	h.Claim("Set-Cookie")
}

func TestClaimClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Claim("Foo-Key")
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf(`h.Claim("Foo-Key") expected panic`)
	}()
	h.Claim("Foo-Key")
}

func TestHeaderIsClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Claim("Foo-Key")
	if got := h.IsClaimed("Foo-Key"); got != true {
		t.Errorf(`h.IsClaimed("Foo-Key") got: %v want: true`, got)
	}
}

func TestHeaderIsClaimedCanonicalization(t *testing.T) {
	h := NewHeader(http.Header{})
	h.Claim("fOo-KEY")
	if got := h.IsClaimed("foo-keY"); got != true {
		t.Errorf(`h.IsClaimed("foo-keY") got: %v want: true`, got)
	}
}

func TestHeaderIsClaimedNotClaimed(t *testing.T) {
	h := NewHeader(http.Header{})
	if got := h.IsClaimed("Foo-Key"); got != false {
		t.Errorf(`h.IsClaimed("Foo-Key") got: %v want: true`, got)
	}
}

func TestHeaderIsClaimedSetCookie(t *testing.T) {
	h := NewHeader(http.Header{})
	if got := h.IsClaimed("Set-Cookie"); got != true {
		t.Errorf(`h.IsClaimed("Set-Cookie") got: %v want: true`, got)
	}
}
