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
			want: map[string][]string{"A": []string{"B"}},
		},
		{
			name: "Ordering",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: X\r\n" +
				"A: Y\r\n" +
				"\r\n"),
			want: map[string][]string{"A": []string{"X", "Y"}},
		},
		{
			name: "Casing",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"BaBaBaBa-BaBaBaBa-BaBaBa: xXxXxXxX\r\n" +
				"cDcDcDcD-cDcDcDcD-cDcDcD: YyYyYyYy\r\n" +
				"\r\n"),
			want: map[string][]string{
				"Babababa-Babababa-Bababa": []string{"xXxXxXxX"},
				"Cdcdcdcd-Cdcdcdcd-Cdcdcd": []string{"YyYyYyYy"},
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
			want: map[string][]string{"A": []string{"X", "Y", "Z", "W"}},
		},
		{
			name: "MultiLineHeaderSpace",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AAAA: aaaa aaa\r\n" +
				" aaa aaa\r\n" +
				"\r\n"),
			want: http.Header{"Aaaa": []string{"aaaa aaa aaa aaa"}},
		},
		{
			name: "MultiLineHeaderTab",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AAAA: aaaa aaa\r\n" +
				"\taaa aaa\r\n" +
				"\r\n"),
			want: http.Header{"Aaaa": []string{"aaaa aaa aaa aaa"}},
		},
		{
			name: "MultiLineHeaderManyLines",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AAAA: aaaa aaa\r\n" +
				" aaa\r\n" +
				" aaa\r\n" +
				"\r\n"),
			want: http.Header{"Aaaa": []string{"aaaa aaa aaa aaa"}},
		},
		{
			name: "UnicodeValue",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A: \xf0\x9f\xa5\xb3\r\n" +
				"\r\n"),
			want: http.Header{"A": []string{"\xf0\x9f\xa5\xb3"}},
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

			if got, want := extractStatus(resp), statusOK; got != want {
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
			want: statusBadRequest,
		},
		{
			name: "MultilineHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AA\r\n" +
				" AA: aaaa\r\n" +
				"\r\n"),
			want: statusBadRequest,
		},
		{
			name: "MultilineHeaderBeforeColon",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A\r\n" +
				" : a\r\n" +
				"\r\n"),
			want: statusBadRequest,
		},
		{
			name: "UnicodeHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\xf0\x9f\xa5\xb3: a\r\n" +
				"\r\n"),
			want: statusInvalidHeaderName,
		},
		{
			name: "SpecialCharactersHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"&%*(#@%()): a\r\n" +
				"\r\n"),
			want: statusInvalidHeaderName,
		},
		{
			name: "NullByteInHeaderName",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"A\x00A: a\r\n" +
				"\r\n"),
			want: statusInvalidHeaderName,
		},
		{
			name: "NullByteInHeaderValue",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"AA: a\x00a\r\n" +
				"\r\n"),
			want: statusInvalidHeaderValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, nil)
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got := extractStatus(resp); got != tt.want {
				t.Errorf("status code got: %q want: %q", got, tt.want)
			}
		})
	}
}

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

			if got, want := extractStatus(resp), statusOK; got != want {
				t.Errorf("status code got: %q want: %q", got, want)
			}
		})
	}
}
