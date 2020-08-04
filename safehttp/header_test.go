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
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
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
	h := newHeader(http.Header{})
	if err := h.Set("fOo-KeY", "Bar-Value"); err != nil {
		t.Errorf(`h.Set("fOo-KeY", "Bar-Value") got err: %v want: nil`, err)
	}
	if got, want := h.Get("FoO-kEy"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("FoO-kEy") got: %q want %q`, got, want)
	}
}

func TestSetSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	cookie := &http.Cookie{Name: "x", Value: "y"}
	h.SetCookie(cookie)
	if err := h.Set("Set-Cookie", "x=y"); err == nil {
		t.Error(`h.Set("Set-Cookie", "x=y") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"x=y"}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}
func TestSetEmptySetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Set-Cookie", "x=y"); err == nil {
		t.Error(`h.Set("Set-Cookie", "x=y") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestSetClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Pizza-Value"); err != nil {
		t.Errorf(`h.Set("Foo-Key", "Pizza-Value") got: %v want: nil`, err)
	}
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if err := h.Set("Foo-Key", "Bar-Value"); err == nil {
		t.Error(`h.Set("Foo-Key", "Bar-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"Pizza-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestSetEmptyClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if err := h.Set("Foo-Key", "Bar-Value"); err == nil {
		t.Error(`h.Set("Foo-Key", "Bar-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestAdd(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Add("Foo-Key", "Pizza-Value"); err != nil {
		t.Errorf(`h.Add("Foo-Key", "Pizza-Value") got err: %v want: nil`, err)
	}
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
	h := newHeader(http.Header{})
	if err := h.Add("fOo-KeY", "Bar-Value"); err != nil {
		t.Errorf(`h.Add("fOo-KeY", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Add("FoO-kEy", "Pizza-Value"); err != nil {
		t.Errorf(`h.Add("FoO-kEy", "Pizza-Value") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{"Bar-Value", "Pizza-Value"}, h.Values("fOO-KEY")); diff != "" {
		t.Errorf("h.Values(\"fOO-KEY\")) mismatch (-want +got):\n%s", diff)
	}
}

func TestAddSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	cookie := &http.Cookie{Name: "x", Value: "y"}
	h.SetCookie(cookie)
	if err := h.Add("Set-Cookie", "x=y"); err == nil {
		t.Error(`h.Add("Set-Cookie", "x=y") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"x=y"}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}
func TestAddEmptySetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Set-Cookie", "x=y"); err == nil {
		t.Error(`h.Add("Set-Cookie", "x=y") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestAddClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if err := h.Add("Foo-Key", "Pizza-Value"); err == nil {
		t.Error(`h.Add("Foo-Key", "Pizza-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestAddEmptyClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if err := h.Add("Foo-Key", "Pizza-Value"); err == nil {
		t.Error(`h.Add("Foo-Key", "Pizza-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDel(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Del("Foo-Key"); err != nil {
		t.Errorf(`h.Del("Foo-Key") got err: %v want: nil`, err)
	}
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
	h := newHeader(http.Header{})
	if err := h.Set("fOo-KeY", "Bar-Value"); err != nil {
		t.Errorf(`h.Set("fOo-KeY", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Del("FoO-kEy"); err != nil {
		t.Errorf(`h.Del("FoO-kEy") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{}, h.Values("FOO-kEY")); diff != "" {
		t.Errorf("h.Values(\"FOO-kEY\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	cookie := &http.Cookie{Name: "x", Value: "y"}
	h.SetCookie(cookie)
	if err := h.Del("Set-Cookie"); err == nil {
		t.Error(`h.Del("Set-Cookie") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"x=y"}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelEmptySetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Del("Set-Cookie"); err == nil {
		t.Error(`h.Del("Set-Cookie") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf("h.Values(\"Set-Cookie\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if err := h.Del("Foo-Key"); err == nil {
		t.Error(`h.Del("Foo-Key") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestDelEmptyClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if err := h.Del("Foo-Key"); err == nil {
		t.Error(`h.Del("Foo-Key") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	c := &http.Cookie{Name: "x", Value: "y"}
	h.SetCookie(c)
	if got, want := h.Get("Set-Cookie"), "x=y"; got != want {
		t.Errorf(`h.Get("Set-Cookie") got: %q want: %q`, got, want)
	}
}

func TestSetCookieInvalidName(t *testing.T) {
	h := newHeader(http.Header{})
	c := &http.Cookie{Name: "x=", Value: "y"}
	h.SetCookie(c)
	if got, want := h.Get("Set-Cookie"), ""; got != want {
		t.Errorf(`h.Get("Set-Cookie") got: %q want: %q`, got, want)
	}
}

// TestValuesModifyClaimed verifies that modifying the
// slice returned by Values() doesn't modify the underlying
// slice. The test ensures that Values() returns a copy
// of the underlying slice.
func TestValuesModifyClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`h.Claim("Foo-Key") got: %v want: nil`, err)
	}
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
			h := newHeader(http.Header{})
			if err := h.Add("Foo-Key", tt.values[0]); err != nil {
				t.Errorf(`h.Add("Foo-Key", tt.values[0]) got err: %v want: nil`, err)
			}
			if err := h.Add("Foo-Key", tt.values[1]); err != nil {
				t.Errorf(`h.Add("Foo-Key", tt.values[1]) got err: %v want: nil`, err)
			}
			if diff := cmp.Diff(tt.values, h.Values("Foo-Key")); diff != "" {
				t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestManyEqualKeyValuePairs(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{"Bar-Value", "Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

// TestAddSet tests that calling Set() after calling Add() will
// correctly remove the previously added header before setting
// the new one.
func TestAddSet(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Errorf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Set("Foo-Key", "Pizza-Value"); err != nil {
		t.Errorf(`h.Set("Foo-Key", "Pizza-Value") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{"Pizza-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestValuesEmptyHeader(t *testing.T) {
	h := newHeader(http.Header{})
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf("h.Values(\"Foo-Key\") mismatch (-want +got):\n%s", diff)
	}
}

func TestClaim(t *testing.T) {
	h := newHeader(http.Header{})
	set, err := h.Claim("Foo-Key")
	if err != nil {
		t.Fatalf(`_, err := h.Claim("Foo-Key") got: %v want: nil`, err)
	}
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
	h := newHeader(http.Header{})
	set, err := h.Claim("fOO-kEY")
	if err != nil {
		t.Fatalf(`_, err := h.Claim("fOO-kEY") got: %v want: nil`, err)
	}
	set([]string{"Bar-Value", "Pizza-Value"})
	if diff := cmp.Diff([]string{"Bar-Value", "Pizza-Value"}, h.Values("fOo-kEy")); diff != "" {
		t.Errorf("h.Values(\"fOo-kEy\") mismatch (-want +got):\n%s", diff)
	}
}

func TestClaimSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if _, err := h.Claim("Set-Cookie"); err == nil {
		t.Error(`_, err := h.Claim("Set-Cookie") got: nil want: error`)
	}
}

func TestClaimClaimed(t *testing.T) {
	h := newHeader(http.Header{})
	if _, err := h.Claim("Foo-Key"); err != nil {
		t.Errorf(`_, err := h.Claim("Foo-Key") got: %v want: nil`, err)
	}
	if _, err := h.Claim("Foo-Key"); err == nil {
		t.Error(`_, err := h.Claim("Foo-Key") got: nil want: error`)
	}
}
