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

package staticheaders_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/staticheaders"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestPlugin(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	rr := safehttptest.NewResponseRecorder()

	p := staticheaders.Plugin{}
	p.Before(rr.ResponseWriter, req)

	if got, want := rr.Status(), 200; got != want {
		t.Errorf("rr.Status() got: %v want: %v", got, want)
	}

	wantHeaders := map[string][]string{
		"X-Content-Type-Options": {"nosniff"},
		"X-Xss-Protection":       {"0"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
		t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
	}

	if got, want := rr.Body(), ""; got != want {
		t.Errorf("rr.Body() got: %q want: %q", got, want)
	}
}

func TestAlreadyClaimed(t *testing.T) {
	alreadyClaimed := []string{"X-Content-Type-Options", "X-XSS-Protection"}

	for _, h := range alreadyClaimed {
		t.Run(h, func(t *testing.T) {
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
			rr := safehttptest.NewResponseRecorder()
			if _, err := rr.ResponseWriter.Header().Claim(h); err != nil {
				t.Fatalf("rr.ResponseWriter.Header().Claim(h) got: %v want: nil", err)
			}

			p := staticheaders.Plugin{}
			p.Before(rr.ResponseWriter, req)

			if got, want := rr.Status(), 500; got != want {
				t.Errorf("rr.Status() got: %v want: %v", got, want)
			}

			wantHeaders := map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			}
			if diff := cmp.Diff(wantHeaders, map[string][]string(rr.Header())); diff != "" {
				t.Errorf("rr.Header() mismatch (-want +got):\n%s", diff)
			}

			if got, want := rr.Body(), "Internal Server Error\n"; got != want {
				t.Errorf("rr.Body() got: %q want: %q", got, want)
			}
		})
	}
}
