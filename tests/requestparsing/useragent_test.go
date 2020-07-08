package requestparsing

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/testing/requesttesting"
)

func TestUserAgent(t *testing.T) {
	type userAgentWant struct {
		headers   map[string][]string
		useragent string
	}

	var userAgentTests = []struct {
		name    string
		request []byte
		want    userAgentWant
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"User-Agent: BlahBlah\r\n" +
				"\r\n"),
			want: userAgentWant{
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
			want: userAgentWant{
				headers:   map[string][]string{"User-Agent": []string{"BlahBlah", "FooFoo"}},
				useragent: "BlahBlah",
			},
		},
	}

	for _, tt := range userAgentTests {
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

			if !bytes.HasPrefix(resp, []byte(statusOK)) {
				got := string(resp[:bytes.IndexByte(resp, '\n')+1])
				t.Errorf("status code got: %q want: %q", got, statusOK)
			}
		})
	}
}
