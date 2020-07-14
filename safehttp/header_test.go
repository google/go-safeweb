package safehttp

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSet(t *testing.T) {
	h := NewHeader()
	if err := h.Set("pizza-pasta", "potato-carrot"); err != nil {
		t.Fatalf(`h.Set("pizza-pasta", "potato-carrot") got err: %v`, err)
	}
	if got, want := h.Get("pIzza-pAsta"), "potato-carrot"; got != want {
		t.Errorf(`h.Get("pIzza-pAsta") got: %q want %q`, got, want)
	}
}

func TestSetDisallowed(t *testing.T) {
	h := NewHeader()
	err := h.Set("Set-Cookie", "x=y")
	if got, want := err.Error(), `The header with name "Set-Cookie" is disallowed.`; got != want {
		t.Errorf(`h.Set("Set-Cookie", "x=y") got: %v want: %v`, got, want)
	}
	if diff := cmp.Diff([]string(nil), h.Values("set-cookie")); diff != "" {
		t.Errorf(`h.Values("set-cookie") mismatch (-want +got):\n%s`, diff)
	}
}

func TestSetImmutable(t *testing.T) {
	h := NewHeader()
	h.MarkImmutable("pizza-pasta")
	err := h.Set("pizza-pasta", "potato-carrot")
	if got, want := err.Error(), `The header with name "Pizza-Pasta" is immutable.`; got != want {
		t.Errorf(`h.Set("pizza-pasta", "potato-carrot") got: %v want: %v`, got, want)
	}
	if diff := cmp.Diff([]string(nil), h.Values("pizza-pasta")); diff != "" {
		t.Errorf(`h.Values("pizza-pasta") mismatch (-want +got):\n%s`, diff)
	}
}

func TestAdd(t *testing.T) {
	h := NewHeader()
	if err := h.Add("pizza-pasta", "potato-carrot"); err != nil {
		t.Fatalf(`h.Add("pizza-pasta", "potato-carrot") got err: %v`, err)
	}
	if err := h.Add("pizzA-pastA", "banana-apple"); err != nil {
		t.Fatalf(`h.Add("pizzA-pastA", "banana-apple") got err: %v`, err)
	}
	if diff := cmp.Diff([]string{"potato-carrot", "banana-apple"}, h.Values("pizza-pasta")); diff != "" {
		t.Errorf(`h.Values("pizza-pasta") mismatch (-want +got):\n%s`, diff)
	}
}

func TestAddDisallowed(t *testing.T) {
	h := NewHeader()
	err := h.Add("Set-Cookie", "potato-carrot")
	if got, want := err.Error(), `The header with name "Set-Cookie" is disallowed.`; got != want {
		t.Errorf(`h.Add("Set-Cookie", "potato-carrot") got: %v want: %v`, got, want)
	}
	if diff := cmp.Diff([]string(nil), h.Values("set-cookie")); diff != "" {
		t.Errorf(`h.Values("set-cookie") mismatch (-want +got):\n%s`, diff)
	}
}

func TestAddImmutable(t *testing.T) {
	h := NewHeader()
	if err := h.Add("pizza-paSta", "potato-carrot"); err != nil {
		t.Fatalf(`h.Add("pizza-paSta", "potato-carrot") got err: %v`, err)
	}
	h.MarkImmutable("pIzza-pasta")
	err := h.Add("pizza-paSta", "banana-apple")
	if got, want := err.Error(), `The header with name "Pizza-Pasta" is immutable.`; got != want {
		t.Errorf(`h.Add("Set-Cookie", "potato-carrot") got: %v want: %v`, got, want)
	}
	if diff := cmp.Diff([]string{"potato-carrot"}, h.Values("pizza-pasta")); diff != "" {
		t.Errorf(`h.Values("pizza-pasta") mismatch (-want +got):\n%s`, diff)
	}
}

func TestDel(t *testing.T) {
	h := NewHeader()
	if err := h.Set("piZza-pasTa", "potato-carrot"); err != nil {
		t.Fatalf(`h.Set("piZza-pasTa", "potato-carrot") got err: %v`, err)
	}
	if err := h.Del("piZZa-pasta"); err != nil {
		t.Fatalf(`h.Del("piZZa-pasta") got err: %v`, err)
	}
	if diff := cmp.Diff([]string(nil), h.Values("pizza-pasta")); diff != "" {
		t.Errorf(`h.Values("pizza-pasta") mismatch (-want +got):\n%s`, diff)
	}
}

func TestDelDisallowed(t *testing.T) {
	h := NewHeader()
	err := h.Del("Set-Cookie")
	if got, want := err.Error(), `The header with name "Set-Cookie" is disallowed.`; got != want {
		t.Errorf(`h.Del("Set-Cookie") got: %v want: %v`, got, want)
	}
}

func TestDelImmutable(t *testing.T) {
	h := NewHeader()
	if err := h.Set("piZza-pasTa", "potato-carrot"); err != nil {
		t.Fatalf(`h.Set("piZza-pasTa", "potato-carrot") got err: %v`, err)
	}
	h.MarkImmutable("PIZZA-PASTA")
	err := h.Del("piZZa-pasta")
	if got, want := err.Error(), `The header with name "Pizza-Pasta" is immutable.`; got != want {
		t.Errorf(`h.Add("Set-Cookie", "potato-carrot") got: %v want: %v`, got, want)
	}
	if diff := cmp.Diff([]string{"potato-carrot"}, h.Values("pizza-pasta")); diff != "" {
		t.Errorf(`h.Values("pizza-pasta") mismatch (-want +got):\n%s`, diff)
	}
}

func TestWriteSubSet(t *testing.T) {
	h := NewHeader()
	if err := h.Set("pizzA-pastA", "potato-carrot"); err != nil {
		t.Fatalf(`h.Set("pizzA-pastA", "potato-carrot") got err: %v`, err)
	}
	if err := h.Add("pIzza-pAsta", "orange-pear"); err != nil {
		t.Fatalf(`h.Add("pIzza-pAsta", "orange-pear") got err: %v`, err)
	}
	if err := h.Add("bolognESE-LASAGna", "tomato-apple"); err != nil {
		t.Fatalf(`h.Add("bolognESE-LASAGna", "tomato-apple") got err: %v`, err)
	}
	if err := h.Add("tACO-kebab", "lettuce-tomato"); err != nil {
		t.Fatalf(`h.Add("tACO-kebab", "lettuce-tomato") got err: %v`, err)
	}

	exclude := map[string]bool{"PIZZA-pasta": true, "BOLOGNESE-lasagna": true}
	builder := &strings.Builder{}
	if err := h.WriteSubset(builder, exclude); err != nil {
		t.Fatalf(`h.Set("pizzA-pastA", "potato-carrot") got err: %v`, err)
	}

	got := builder.String()
	if !strings.Contains(got, "Taco-Kebab") {
		t.Error(`strings.Contains(got, "Taco-Kebab") got: false want: true`)
	}
	if strings.Contains(got, "Pizza-Pasta") {
		t.Error(`strings.Contains(got, "Pizza-Pasta") got: true want: false`)
	}
	if strings.Contains(got, "Bolognese-Lasagna") {
		t.Error(`strings.Contains(got, "Bolognese-Lasagna") got: true want: false`)
	}
}

func TestSetCookie(t *testing.T) {
	h := NewHeader()
	c := &http.Cookie{Name: "x", Value: "y"}
	h.SetCookie(c)
	if got, want := h.Get("Set-Cookie"), "x=y"; got != want {
		t.Errorf(`h.Get("Set-Cookie") got: %q want: %q`, got, want)
	}
}

func TestSetCookieInvalidName(t *testing.T) {
	h := NewHeader()
	c := &http.Cookie{Name: "x=", Value: "y"}
	h.SetCookie(c)
	if got, want := h.Get("Set-Cookie"), ""; got != want {
		t.Errorf(`h.Get("Set-Cookie") got: %q want: %q`, got, want)
	}
}
