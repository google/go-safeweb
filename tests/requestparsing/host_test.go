package requestparsing

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/go-safeweb/testing/requesttesting"
)

var hostTests = []struct {
	name    string
	request []byte
	want    string
}{
	{
		name: "Basic",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"\r\n"),
		want: "localhost:8080",
	},
	{
		// https://tools.ietf.org/html/rfc7230#section-5.3.2
		name: "AbsoluteFormURL",
		request: []byte("GET http://y.com/asdf HTTP/1.1\r\n" +
			"Host: x.com\r\n" +
			"\r\n"),
		want: "y.com",
	},
	{
		name: "AbsoluteFormURLNoValidSchemaNeeded",
		request: []byte("GET x://y.com/asdf HTTP/1.1\r\n" +
			"Host: x.com\r\n" +
			"\r\n"),
		want: "y.com",
	},
	{
		// https://tools.ietf.org/html/rfc7230#section-5.3.3
		name: "AuthorityForm",
		request: []byte("GET y.com:123/asdf HTTP/1.1\r\n" +
			"Host: x.com\r\n" +
			"\r\n"),
		want: "x.com",
	},
	{
		name: "NoDoubleSlash",
		request: []byte("GET http:y.com/asdf HTTP/1.1\r\n" +
			"Host: x.com\r\n" +
			"\r\n"),
		want: "x.com",
	},
	{
		name: "NoSchemaOnlyDoubleSlash",
		request: []byte("GET //y.com/asdf HTTP/1.1\r\n" +
			"Host: x.com\r\n" +
			"\r\n"),
		want: "x.com",
	},
}

func TestHostHeader(t *testing.T) {
	for _, tt := range hostTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if len(r.Header) != 0 {
					t.Errorf("len(r.Header) got: %v want: 0", len(r.Header))
				}

				if r.Host != tt.want {
					t.Errorf("r.Host got: %q want: %q", r.Host, tt.want)
				}
			})
			if err != nil {
				t.Errorf("MakeRequest() got: %v want: nil", err)
			}

			if !bytes.HasPrefix(resp, []byte(statusOK)) {
				got := string(resp[:bytes.IndexByte(resp, '\n')+1])
				t.Errorf("status code got: %q want: %q", got, statusOK)
			}
		})
	}
}

func TestHostHeaderMultiple(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: x.com\r\n" +
		"Host: y.com\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, nil)
	if err != nil {
		t.Errorf("MakeRequest() got: %v want: nil", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusTooManyHostHeaders)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusTooManyHostHeaders)
	}
}
