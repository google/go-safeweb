package requestparsing

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/google/go-safeweb/testing/requesttesting"
)

func TestReferer(t *testing.T) {
	type refererWant struct {
		headers map[string][]string
		referer string
	}

	var refererTests = []struct {
		name    string
		request []byte
		want    refererWant
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Referer: http://example.com\r\n" +
				"\r\n"),
			want: refererWant{
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
			want: refererWant{
				headers: map[string][]string{"Referer": []string{"http://example.com", "http://evil.com"}},
				referer: "http://example.com",
			},
		},
	}

	for _, tt := range refererTests {
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

			if !bytes.HasPrefix(resp, []byte(statusOK)) {
				got := string(resp[:bytes.IndexByte(resp, '\n')+1])
				t.Errorf("status code got: %q want: %q", got, statusOK)
			}
		})
	}
}
