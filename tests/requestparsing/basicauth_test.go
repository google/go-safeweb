package requestparsing

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/google/go-safeweb/testing/requesttesting"
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
				// Base64 encoding of "Pelle:Password".
				"Authorization: Basic UGVsbGU6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "Pelle",
				password: "Password",
				ok:       true,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"Basic UGVsbGU6UGFzc3dvcmQ="}},
		},
		{
			name: "NoTrailingEquals",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password" without trailing equals.
				"Authorization: Basic UGVsbGU6UGFzc3dvcmQ\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "",
				password: "",
				ok:       false,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"Basic UGVsbGU6UGFzc3dvcmQ"}},
		},
		{
			name: "DoubleColon",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password:Password".
				"Authorization: Basic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "Pelle",
				password: "Password:Password",
				ok:       true,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"Basic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ="}},
		},
		{
			name: "NotBasic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password:Password".
				"Authorization: xasic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			wantBasicAuth: basicAuth{
				username: "",
				password: "",
				ok:       false,
			},
			// Same Base64 as above.
			wantHeaders: map[string][]string{"Authorization": []string{"xasic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ="}},
		},
		{
			name: "Ordering",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "AAA:aaa".
				"Authorization: basic QUFBOmFhYQ==\r\n" +
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

			if !bytes.HasPrefix(resp, []byte(statusOK)) {
				got := string(resp[:bytes.IndexByte(resp, '\n')+1])
				t.Errorf("status code got: %q want: %q", got, statusOK)
			}
		})
	}
}
