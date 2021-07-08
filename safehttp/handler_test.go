package safehttp_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

var BarHandler = safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	if !strings.HasPrefix(r.URL().Path(), "/bar") {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	return w.Write(safehtml.HTMLEscaped("Hello!"))
})

func TestStripPrefix(t *testing.T) {
	mux := safehttp.NewServeMuxConfig(nil).Mux()

	mux.Handle("/bar", safehttp.MethodGet, BarHandler)
	mux.Handle("/more/bar", safehttp.MethodGet, safehttp.StripPrefix("/more", BarHandler))

	r := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/bar", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, r)

	rStrip := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/more/bar", nil)
	rwStrip := httptest.NewRecorder()
	mux.ServeHTTP(rwStrip, rStrip)

	if rwStrip.Code != rw.Code {
		t.Errorf("Code got %v, want %v", rwStrip.Code, rw.Code)
	}

	if diff := cmp.Diff(rw.Header(), rwStrip.Header()); diff != "" {
		t.Errorf("Header() mismatch (-want +got):\n%s", diff)
	}

	if got := rwStrip.Body.String(); got != rw.Body.String() {
		t.Errorf("response body: got %q want %q", got, rw.Body.String())
	}
}

func TestStripPrefixPanic(t *testing.T) {
	mux := safehttp.NewServeMuxConfig(nil).Mux()

	mux.Handle("/bar", safehttp.MethodGet, BarHandler)
	mux.Handle("/more/bar", safehttp.MethodGet, safehttp.StripPrefix("/badprefix", BarHandler))

	r := httptest.NewRequest(safehttp.MethodGet, "http://foo.com/more/bar", nil)
	rw := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf("expected panic")
	}()
	mux.ServeHTTP(rw, r)
}
