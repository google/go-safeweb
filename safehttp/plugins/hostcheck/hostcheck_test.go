package hostcheck_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/hostcheck"
	"github.com/google/go-safeweb/safehttp/safehttptest"
	"github.com/google/safehtml"
)

func TestInterceptor(t *testing.T) {
	var test = []struct {
		name       string
		req        *http.Request
		wantStatus safehttp.StatusCode
	}{
		{
			name:       "Valid Host",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://foo.com/", nil),
			wantStatus: safehttp.StatusOK,
		},
		{
			name:       "Invalid Host",
			req:        httptest.NewRequest(safehttp.MethodGet, "http://bar.com/", nil),
			wantStatus: safehttp.StatusNotFound,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mux := safehttp.NewServeMux(safehttp.DefaultDispatcher{})
			mux.Install(hostcheck.New("foo.com"))

			h := safehttp.HandlerFunc(func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
				return w.Write(safehtml.HTMLEscaped("<h1>Hello World!</h1>"))
			})
			mux.Handle("/", safehttp.MethodGet, h)

			b := &strings.Builder{}
			rw := safehttptest.NewTestResponseWriter(b)
			mux.ServeHTTP(rw, tt.req)

			if rw.Status() != tt.wantStatus {
				t.Errorf("rw.Status(): got %v want %v", rw.Status(), tt.wantStatus)
			}
		})
	}
}
