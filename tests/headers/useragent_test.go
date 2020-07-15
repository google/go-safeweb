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

func TestUserAgent(t *testing.T) {
	type testWant struct {
		headers   map[string][]string
		useragent string
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
				"User-Agent: BlahBlah\r\n" +
				"\r\n"),
			want: testWant{
				headers:   map[string][]string{"User-Agent": []string{"BlahBlah"}},
				useragent: "BlahBlah",
			},
		},
		{
			name: "Ordering",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"User-Agent: BlahBlah\r\n" +
				"User-Agent: FooFoo\r\n" +
				"\r\n"),
			want: testWant{
				headers:   map[string][]string{"User-Agent": []string{"BlahBlah", "FooFoo"}},
				useragent: "BlahBlah",
			},
		},
		{
			name: "CasingOrdering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"user-Agent: BlahBlah\r\n" +
				"User-Agent: FooFoo\r\n" +
				"\r\n"),
			want: testWant{
				headers:   map[string][]string{"User-Agent": []string{"BlahBlah", "FooFoo"}},
				useragent: "BlahBlah",
			},
		},
		{
			name: "CasingOrdering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"User-Agent: BlahBlah\r\n" +
				"user-Agent: FooFoo\r\n" +
				"\r\n"),
			want: testWant{
				headers:   map[string][]string{"User-Agent": []string{"BlahBlah", "FooFoo"}},
				useragent: "BlahBlah",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want.headers, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}

				if r.UserAgent() != tt.want.useragent {
					t.Errorf("r.UserAgent() got: %q want: %q", r.UserAgent(), tt.want.useragent)
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
