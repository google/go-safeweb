package internal_test

import (
	"net/http"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/internal"
)

func TestRawRequest(t *testing.T) {
	if _, ok := internal.RawRequest.(func(*safehttp.IncomingRequest) *http.Request); !ok {
		t.Errorf("RawRequest type got %T, want func(*safehttp.IncomingRequest) *http.Request", internal.RawRequest)
	}
}
