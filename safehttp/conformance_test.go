package safehttp_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
)

func TestConformance(t *testing.T) {
	m := safehttp.NewServeMux(nil)
	m.InstallConformanceCheck("myCheck", func(pattern, method string, interceps []safehttp.ConfiguredInterceptor) error {
		return errors.New("failed")
	})
	m.Handle("/foo", safehttp.MethodPost, safehttp.HandlerFunc(
		func(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
			t.Fatal("handler should not be reached")
			panic("unreachable")
		}))

	defer func() {
		// We're testing the panic message a bit, to find whether the right
		// things get propagated. This is crucial for user experience.
		if r := recover(); r == nil {
			t.Fatal("expected a panic")
		} else if !strings.Contains(fmt.Sprintf("%v", r), "myCheck") {
			t.Fatal(`expected panic message to contain "myCheck"`)
		} else if !strings.Contains(fmt.Sprintf("%v", r), "conformance") {
			t.Fatal(`expected panic message to contain "conformance"`)
		}
	}()

	m.ServeHTTP(nil, nil)
}
