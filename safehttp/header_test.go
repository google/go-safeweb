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
		t.Fatalf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if got, want := h.Get("Foo-Key"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("Foo-Key") got: %q want %q`, got, want)
	}
}

func TestSetCanonicalization(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("fOo-KeY", "Bar-Value"); err != nil {
		t.Fatalf(`h.Set("fOo-KeY", "Bar-Value") got err: %v want: nil`, err)
	}
	if got, want := h.Get("FoO-kEy"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("FoO-kEy") got: %q want %q`, got, want)
	}
}

func TestSetSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Set-Cookie", "x=y"); err == nil {
		t.Error(`h.Set("Set-Cookie", "x=y") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf(`h.Values("Set-Cookie") mismatch (-want +got):\n%s`, diff)
	}
}

func TestSetImmutable(t *testing.T) {
	h := newHeader(http.Header{})
	h.MarkImmutable("Foo-Key")
	if err := h.Set("Foo-Key", "Bar-Value"); err == nil {
		t.Error(`h.Set("Foo-Key", "Bar-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf(`h.Values("Foo-Key") mismatch (-want +got):\n%s`, diff)
	}
}

func TestAdd(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Fatalf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Add("Foo-Key", "Bar-Value-2"); err != nil {
		t.Fatalf(`h.Add("Foo-Key", "Bar-Value-2") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{"Bar-Value", "Bar-Value-2"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf(`h.Values("Foo-Key") mismatch (-want +got):\n%s`, diff)
	}
}

func TestAddCanonicalization(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("fOo-KeY", "Bar-Value"); err != nil {
		t.Fatalf(`h.Add("fOo-KeY", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Add("FoO-kEy", "Bar-Value-2"); err != nil {
		t.Fatalf(`h.Add("FoO-kEy", "Bar-Value-2") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{"Bar-Value", "Bar-Value-2"}, h.Values("fOO-KEY")); diff != "" {
		t.Errorf(`h.Values("fOO-KEY")) mismatch (-want +got):\n%s`, diff)
	}
}

func TestAddSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Set-Cookie", "x=y"); err == nil {
		t.Error(`h.Add("Set-Cookie", "x=y") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{}, h.Values("Set-Cookie")); diff != "" {
		t.Errorf(`h.Values("Set-Cookie") mismatch (-want +got):\n%s`, diff)
	}
}

func TestAddImmutable(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Add("Foo-Key", "Bar-Value"); err != nil {
		t.Fatalf(`h.Add("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	h.MarkImmutable("Foo-Key")
	if err := h.Add("Foo-Key", "Bar-Value-2"); err == nil {
		t.Error(`h.Add("Set-Cookie", "Bar-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf(`h.Values("Foo-Key") mismatch (-want +got):\n%s`, diff)
	}
}

func TestDel(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Fatalf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Del("Foo-Key"); err != nil {
		t.Fatalf(`h.Del("Foo-Key") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{}, h.Values("Foo-Key")); diff != "" {
		t.Errorf(`h.Values("Foo-Key") mismatch (-want +got):\n%s`, diff)
	}
}

func TestDelCanonicalization(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("fOo-KeY", "Bar-Value"); err != nil {
		t.Fatalf(`h.Set("fOo-KeY", "Bar-Value") got err: %v want: nil`, err)
	}
	if err := h.Del("FoO-kEy"); err != nil {
		t.Fatalf(`h.Del("FoO-kEy") got err: %v want: nil`, err)
	}
	if diff := cmp.Diff([]string{}, h.Values("FOO-kEY")); diff != "" {
		t.Errorf(`h.Values("FOO-kEY") mismatch (-want +got):\n%s`, diff)
	}
}

func TestDelSetCookie(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Del("Set-Cookie"); err == nil {
		t.Error(`h.Del("Set-Cookie") got: nil want: error`)
	}
}

func TestDelImmutable(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Fatalf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	h.MarkImmutable("Foo-Key")
	if err := h.Del("Foo-Key"); err == nil {
		t.Error(`h.Add("Set-Cookie", "Bar-Value") got: nil want: error`)
	}
	if diff := cmp.Diff([]string{"Bar-Value"}, h.Values("Foo-Key")); diff != "" {
		t.Errorf(`h.Values("Foo-Key") mismatch (-want +got):\n%s`, diff)
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

func TestValuesModifyImmutable(t *testing.T) {
	h := newHeader(http.Header{})
	if err := h.Set("Foo-Key", "Bar-Value"); err != nil {
		t.Fatalf(`h.Set("Foo-Key", "Bar-Value") got err: %v want: nil`, err)
	}
	h.MarkImmutable("Foo-Key")
	v := h.Values("Foo-Key")
	if diff := cmp.Diff([]string{"Bar-Value"}, v); diff != "" {
		t.Errorf(`h.Values("Foo-Key") mismatch (-want +got):\n%s`, diff)
	}
	v[0] = "Evil-Value"
	if got, want := h.Get("Foo-Key"), "Bar-Value"; got != want {
		t.Errorf(`h.Get("Foo-Key") got: %v want: %v`, got, want)
	}
}
