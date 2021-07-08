package safehttp

import (
	"github.com/google/go-safeweb/safehttp/internal"
)

func init() {
	internal.RawRequest = rawRequest
}
