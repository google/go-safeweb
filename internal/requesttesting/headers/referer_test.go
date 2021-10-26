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
				headers: map[string][]string{"Referer": {"http://example.com"}},
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
				headers: map[string][]string{"Referer": {"http://example.com", "http://evil.com"}},
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
				headers: map[string][]string{"Referer": {"http://example.com", "http://evil.com"}},
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

func TestRefererOrdering(t *testing.T) {
	// The documentation of http.Request.Referer() doesn't clearly specify
	// that only the first Referer header is used and that the other ones
	// are ignored. This could potentially lead to security issues if two
	// HTTP servers that look at different headers are chained together.
	//
	// The desired behavior would be to respond with 400 (Bad Request)
	// when there is more than one Referer header.

	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Referer: http://example.com\r\n" +
		"Referer: http://evil.com\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			wantHeaders := map[string][]string{"Referer": {"http://example.com", "http://evil.com"}}
			if diff := cmp.Diff(wantHeaders, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}

			if want := "http://example.com"; r.Referer() != want {
				t.Errorf("r.Referer() got: %q want: %q", r.Referer(), want)
			}
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v want: nil", err)
		}

		if got, want := extractStatus(resp), statusOK; got != want {
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

		if got, want := extractStatus(resp), statusBadRequest; got != want {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})
}
