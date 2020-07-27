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

func TestBasicAuth(t *testing.T) {
	type basicAuth struct {
		username string
		password string
		ok       bool
	}

	var tests = []struct {
		name          string
		request       []byte
		wantBasicAuth basicAuth
		wantHeaders   map[string][]string
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pizza:Password".
				"Authorization: Basic UGl6emE6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "Pizza",
				password: "Password",
				ok:       true,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"Basic UGl6emE6UGFzc3dvcmQ="}},
		},
		{
			name: "NoTrailingEquals",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pizza:Password" without trailing equals.
				"Authorization: Basic UGl6emE6UGFzc3dvcmQ\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "",
				password: "",
				ok:       false,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"Basic UGl6emE6UGFzc3dvcmQ"}},
		},
		{
			name: "DoubleColon",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pizza:Password:Password".
				"Authorization: Basic UGl6emE6UGFzc3dvcmQ6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "Pizza",
				password: "Password:Password",
				ok:       true,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"Basic UGl6emE6UGFzc3dvcmQ6UGFzc3dvcmQ="}},
		},
		{
			name: "NotBasic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pizza:Password:Password".
				"Authorization: xasic UGl6emE6UGFzc3dvcmQ6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "",
				password: "",
				ok:       false,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"xasic UGl6emE6UGFzc3dvcmQ6UGFzc3dvcmQ="}},
		},
		{
			name: "CasingOrdering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "AAA:aaa".
				"Authorization: basic QUFBOmFhYQ==\r\n" +
				// Base64 encoding of "BBB:bbb".
				"authorization: basic QkJCOmJiYg==\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "AAA",
				password: "aaa",
				ok:       true,
			},
			// Base64 encoding of "AAA:aaa" and then of "BBB:bbb" in that order.
			wantHeaders: map[string][]string{"Authorization": []string{"basic QUFBOmFhYQ==", "basic QkJCOmJiYg=="}},
		},
		{
			name: "CasingOrdering2",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "AAA:aaa".
				"authorization: basic QUFBOmFhYQ==\r\n" +
				// Base64 encoding of "BBB:bbb".
				"Authorization: basic QkJCOmJiYg==\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "AAA",
				password: "aaa",
				ok:       true,
			},
			// Base64 encoding of "AAA:aaa" and then of "BBB:bbb" in that order.
			wantHeaders: map[string][]string{"Authorization": []string{"basic QUFBOmFhYQ==", "basic QkJCOmJiYg=="}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.wantHeaders, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}

				username, password, ok := r.BasicAuth()
				if ok != tt.wantBasicAuth.ok {
					t.Errorf("_, _, ok := r.BasicAuth() got: %v want: %v", ok, tt.wantBasicAuth.ok)
				}

				if username != tt.wantBasicAuth.username {
					t.Errorf("username, _, _ := r.BasicAuth() got: %q want: %q", username, tt.wantBasicAuth.username)
				}

				if password != tt.wantBasicAuth.password {
					t.Errorf("_, password, _ := r.BasicAuth() got: %q want: %q", password, tt.wantBasicAuth.password)
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

func TestBasicAuthOrdering(t *testing.T) {
	// It is not clear from the documentation of http.Request.BasicAuth()
	// that only the first Authorization header is used and any other ones
	// are ignored. This could potentially lead to security issues if two
	// HTTP servers are chained that look at different headers.
	//
	// The desired behavior would instead be for http.Request.BasicAuth() to
	// return ok as false when there are more than one Authorization header.

	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		// Base64 encoding of "AAA:aaa".
		"Authorization: basic QUFBOmFhYQ==\r\n" +
		// Base64 encoding of "BBB:bbb".
		"Authorization: basic QkJCOmJiYg==\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			// Base64 encoding of "AAA:aaa" and then of "BBB:bbb" in that order.
			wantHeaders := map[string][]string{"Authorization": []string{"basic QUFBOmFhYQ==", "basic QkJCOmJiYg=="}}
			if diff := cmp.Diff(wantHeaders, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}

			username, password, ok := r.BasicAuth()
			if want := true; ok != want {
				t.Errorf("_, _, ok := r.BasicAuth() got: %v want: %v", ok, want)
			}

			if want := "AAA"; username != want {
				t.Errorf("username, _, _ := r.BasicAuth() got: %q want: %q", username, want)
			}

			if want := "aaa"; password != want {
				t.Errorf("_, password, _ := r.BasicAuth() got: %q want: %q", password, want)
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
		_, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			// Base64 encoding of "AAA:aaa" and then of "BBB:bbb" in that order.
			wantHeaders := map[string][]string{"Authorization": []string{"basic QUFBOmFhYQ==", "basic QkJCOmJiYg=="}}
			if diff := cmp.Diff(wantHeaders, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}

			_, _, ok := r.BasicAuth()
			if want := false; ok != want {
				t.Errorf("_, _, ok := r.BasicAuth() got: %v want: %v", ok, want)
			}
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v want: nil", err)
		}
	})
}
