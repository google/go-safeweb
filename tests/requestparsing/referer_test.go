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

package requestparsing

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/google/go-safeweb/testing/requesttesting"
)

func TestReferer(t *testing.T) {
	type testWant struct {
		headers map[string][]string
		referer string
	}

	var tests = []struct {
		name    string
		request []byte
		want    testWant
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Referer: http://example.com\r\n" +
				"\r\n"),
			want: testWant{
				headers: map[string][]string{"Referer": []string{"http://example.com"}},
				referer: "http://example.com",
			},
		},
		{
			name: "Ordering",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Referer: http://example.com\r\n" +
				"Referer: http://evil.com\r\n" +
				"\r\n"),
			want: testWant{
				headers: map[string][]string{"Referer": []string{"http://example.com", "http://evil.com"}},
				referer: "http://example.com",
			},
		},
		{
			name: "CasingOrdering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"referer: http://example.com\r\n" +
				"Referer: http://evil.com\r\n" +
				"\r\n"),
			want: testWant{
				headers: map[string][]string{"Referer": []string{"http://example.com", "http://evil.com"}},
				referer: "http://example.com",
			},
		},
		{
			name: "CasingOrdering2",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Referer: http://example.com\r\n" +
				"referer: http://evil.com\r\n" +
				"\r\n"),
			want: testWant{
				headers: map[string][]string{"Referer": []string{"http://example.com", "http://evil.com"}},
				referer: "http://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want.headers, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}

				if r.Referer() != tt.want.referer {
					t.Errorf("r.Referer() got: %q want: %q", r.Referer(), tt.want.referer)
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
