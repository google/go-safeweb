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
	type basicAuthWant struct {
		headers  map[string][]string
		ok       bool
		username string
		password string
	}

	var basicAuthTests = []struct {
		name    string
		request []byte
		want    basicAuthWant
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password".
				"Authorization: Basic UGVsbGU6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			want: basicAuthWant{
				// Same Base64 as above.
				headers:  map[string][]string{"Authorization": []string{"Basic UGVsbGU6UGFzc3dvcmQ="}},
				ok:       true,
				username: "Pelle",
				password: "Password",
			},
		},
		{
			name: "NoTrailingEquals",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password" without trailing equals.
				"Authorization: Basic UGVsbGU6UGFzc3dvcmQ\r\n" +
				"\r\n"),
			want: basicAuthWant{
				// Same Base64 as above.
				headers:  map[string][]string{"Authorization": []string{"Basic UGVsbGU6UGFzc3dvcmQ"}},
				ok:       false,
				username: "",
				password: "",
			},
		},
		{
			name: "DoubleColon",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password:Password".
				"Authorization: Basic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			want: basicAuthWant{
				// Same Base64 as above.
				headers:  map[string][]string{"Authorization": []string{"Basic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ="}},
				ok:       true,
				username: "Pelle",
				password: "Password:Password",
			},
		},
		{
			name: "NotBasic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				// Base64 encoding of "Pelle:Password:Password".
				"Authorization: xasic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ=\r\n" +
				"\r\n"),
			want: basicAuthWant{
				// Same Base64 as above.
				headers:  map[string][]string{"Authorization": []string{"xasic UGVsbGU6UGFzc3dvcmQ6UGFzc3dvcmQ="}},
				ok:       false,
				username: "",
				password: "",
			},
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
			want: basicAuthWant{
				// Base64 encoding of "AAA:aaa" and then of "BBB:bbb" in that order.
				headers:  map[string][]string{"Authorization": []string{"basic QUFBOmFhYQ==", "basic QkJCOmJiYg=="}},
				ok:       true,
				username: "AAA",
				password: "aaa",
			},
		},
	}

	for _, tt := range basicAuthTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want.headers, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}

				username, password, ok := r.BasicAuth()
				if ok != tt.want.ok {
					t.Errorf("_, _, ok := r.BasicAuth() got: %v want: %v", ok, tt.want.ok)
				}

				if username != tt.want.username {
					t.Errorf("username, _, _ := r.BasicAuth() got: %q want: %q", username, tt.want.username)
				}

				if password != tt.want.password {
					t.Errorf("_, password, _ := r.BasicAuth() got: %q want: %q", password, tt.want.password)
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
