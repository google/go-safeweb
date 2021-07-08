package restricted

import (
	"net/http"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/internal"
)

var rawRequest = internal.RawRequest.(func(*safehttp.IncomingRequest) *http.Request)

func RawRequest(r *safehttp.IncomingRequest) *http.Request {
	return rawRequest(r)
}
