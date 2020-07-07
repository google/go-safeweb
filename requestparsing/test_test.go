package requestparsing

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	statusOK = "HTTP/1.1 200 OK\r\n"
)

func TestCase(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"A: B\r\n" +
		"\r\n")
	response, err := makeRequest(context.Background(), request, func(r *http.Request) {
		want := map[string][]string{"A": []string{"B"}}
		if diff := cmp.Diff(want, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}
	})
	if err != nil {
		t.Errorf("makeRequest() got err %v want nil", err)
	}

	if !bytes.HasPrefix(response, []byte(statusOK)) {
		got := string(response[:bytes.IndexByte(response, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}
