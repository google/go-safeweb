// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package headers

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/google/go-safeweb/internal/requesttesting"
)

func TestHeaderParsing(t *testing.T) {
	var tests = []struct {
		name    string
		request []byte
		want    map[string][]string
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: B\r\n" +
				"\r\n"),
			want: map[string][]string{"A": {"B"}},
		},
		{
			name: "Ordering",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: X\r\n" +
				"A: Y\r\n" +
				"\r\n"),
			want: map[string][]string{"A": {"X", "Y"}},
		},
		{
			name: "Casing",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"BaBaBaBa-BaBaBaBa-BaBaBa: xXxXxXxX\r\n" +
				"cDcDcDcD-cDcDcDcD-cDcDcD: YyYyYyYy\r\n" +
				"\r\n"),
			want: map[string][]string{
				"Babababa-Babababa-Bababa": {"xXxXxXxX"},
				"Cdcdcdcd-Cdcdcdcd-Cdcdcd": {"YyYyYyYy"},
			},
		},
		{
			name: "CasingOrdering",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"a: X\r\n" +
				"A: Y\r\n" +
				"a: Z\r\n" +
				"A: W\r\n" +
				"\r\n"),
			want: map[string][]string{"A": {"X", "Y", "Z", "W"}},
		},
		{
			name: "MultiLineHeaderTab",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AAAA: aaaa aaa\r\n" +
				"\taaa aaa\r\n" +
				"\r\n"),
			want: map[string][]string{"Aaaa": {"aaaa aaa aaa aaa"}},
		},
		{
			name: "MultiLineHeaderManyLines",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AAAA: aaaa aaa\r\n" +
				" aaa\r\n" +
				" aaa\r\n" +
				"\r\n"),
			want: map[string][]string{"Aaaa": {"aaaa aaa aaa aaa"}},
		},
		{
			name: "UnicodeValue",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: \xf0\x9f\xa5\xb3\r\n" +
				"\r\n"),
			want: map[string][]string{"A": {"\xf0\x9f\xa5\xb3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got, want := extractStatus(resp), statusOK; !matchStatus(got, want) {
				t.Errorf("status code got: %q want: %q", got, want)
			}
		})
	}
}

func TestStatusCode(t *testing.T) {
	var tests = []struct {
		name    string
		request []byte
		want    string
	}{
		{
			name: "MultilineContinuationOnFirstLine",
			request: []byte("GET / HTTP/1.1\r\n" +
				" A: a\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
		{
			name: "MultilineHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AA\r\n" +
				" AA: aaaa\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
		{
			name: "MultilineHeaderBeforeColon",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A\r\n" +
				" : a\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
		{
			name: "UnicodeHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\xf0\x9f\xa5\xb3: a\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
		{
			name: "SpecialCharactersHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"&%*(#@%()): a\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
		{
			name: "NullByteInHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A\x00A: a\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
		{
			name: "NullByteInHeaderValue",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AA: a\x00a\r\n" +
				"\r\n"),
			want: statusBadRequestPrefix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, nil)
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got := extractStatus(resp); !matchStatus(got, tt.want) {
				t.Errorf("status code got: %q want: %q", got, tt.want)
			}
		})
	}
}

// TestValues verifies that the http.Header.Values() function
// returns the header values in the order that they are sent
// in the request to the server.
func TestValues(t *testing.T) {
	var tests = []struct {
		name    string
		request []byte
		want    []string
	}{
		{
			name: "Ordering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: X\r\n" +
				"A: Y\r\n" +
				"\r\n"),
			want: []string{"X", "Y"},
		},
		{
			name: "Ordering2",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: Y\r\n" +
				"A: X\r\n" +
				"\r\n"),
			want: []string{"Y", "X"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want, r.Header.Values("A")); diff != "" {
					t.Errorf(`r.Header.Values("A") mismatch (-want +got):\n%s`, diff)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v want: nil", err)
			}

			if got, want := extractStatus(resp), statusOK; !matchStatus(got, want) {
				t.Errorf("status code got: %q want: %q", got, want)
			}
		})
	}
}

func TestMultiLineHeader(t *testing.T) {
	// Multiline continuation has been deprecated as of RFC 7230.
	// " Historically, HTTP header field values could be extended over
	//   multiple lines by preceding each extra line with at least one space
	//   or horizontal tab (obs-fold).  This specification deprecates such
	//   line folding [...]
	//    A server that receives an obs-fold in a request message that is not
	//   within a message/http container MUST either reject the message by
	//   sending a 400 (Bad Request), preferably with a representation
	//   explaining that obsolete line folding is unacceptable, or replace
	//   each received obs-fold with one or more SP octets prior to
	//   interpreting the field value or forwarding the message downstream. "
	// - RFC 7230 Section 3.2.4
	//
	// Currently obs-folds are replaced with spaces before the value of the
	// header is interpreted. This is in line with the RFC. But it would
	// be more robust and future proof to drop the support of multiline
	// continuation entirely and instead respond with a 400 (Bad Request)
	// like the RFC also suggests.

	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"AAAA: aaaa aaa\r\n" +
		" aaa aaa\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			want := map[string][]string{"Aaaa": {"aaaa aaa aaa aaa"}}
			if diff := cmp.Diff(want, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v want: nil", err)
		}

		if got, want := extractStatus(resp), statusOK; !matchStatus(got, want) {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})

	t.Run("Desired behavior", func(t *testing.T) {
		t.Skip()
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			t.Error("Expected handler to not be called!")
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v want: nil", err)
		}

		if got, want := extractStatus(resp), statusBadRequestPrefix; !matchStatus(got, want) {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})
}
