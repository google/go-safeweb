package conformance

import (
	"errors"

	"github.com/google/go-safeweb/safehttp"
)

// SingleInterceptorCheck creates a conformance check that passes if there is
// exactly one, properly configured interceptor found.
// The resulting conformance check passes if and only if checker has found exactly one supported
// interceptor and returned no errors.
func SingleInterceptorCheck(checker func(pattern, method string, ip safehttp.ConfiguredInterceptor) (bool, error)) safehttp.ConformanceCheck {
	return func(pattern, method string, interceps []safehttp.ConfiguredInterceptor) error {
		present := false
		for _, ci := range interceps {
			if found, err := checker(pattern, method, ci); found {
				if present {
					return errors.New("multiple interceptors found")
				}
				present = true

				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}
