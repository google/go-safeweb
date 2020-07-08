package requestparsing

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"../../testing/requesttesting"
)

var headerTests = []struct {
	name    string
	request []byte
	want    map[string][]string
}{
	{
		name: "Basic",
		// All tests use verbatim newline characters instead of using multiline strings to ensure that \r and \n end up in exactly the right places.
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"A: B\r\n" +
			"\r\n"),
		want: map[string][]string{"A": []string{"B"}},
	},
	{
		name: "Ordering",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"A: X\r\n" +
			"A: Y\r\n" +
			"\r\n"),
		want: map[string][]string{"A": []string{"X", "Y"}},
	},
	{
		name: "Casing",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"BaBaBaBa-BaBaBaBa-BaBaBa: xXxXxXxX\r\n" +
			"cDcDcDcD-cDcDcDcD-cDcDcD: YyYyYyYy\r\n" +
			"\r\n"),
		want: map[string][]string{
			"Babababa-Babababa-Bababa": []string{"xXxXxXxX"},
			"Cdcdcdcd-Cdcdcdcd-Cdcdcd": []string{"YyYyYyYy"},
		},
	},
	{
		name: "CasingOrdering",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"a: X\r\n" +
			"A: Y\r\n" +
			"a: Z\r\n" +
			"A: W\r\n" +
			"\r\n"),
		want: map[string][]string{"A": []string{"X", "Y", "Z", "W"}},
	},
	{
		name: "MultiLineHeaderSpace",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"AAAA: aaaa aaa\r\n" +
			" aaa aaa\r\n" +
			"\r\n"),
		want: http.Header{"Aaaa": []string{"aaaa aaa aaa aaa"}},
	},
	{
		name: "MultiLineHeaderTab",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"AAAA: aaaa aaa\r\n" +
			"\taaa aaa\r\n" +
			"\r\n"),
		want: http.Header{"Aaaa": []string{"aaaa aaa aaa aaa"}},
	},
	{
		name: "MultiLineHeaderManyLines",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"AAAA: aaaa aaa\r\n" +
			" aaa\r\n" +
			" aaa\r\n" +
			"\r\n"),
		want: http.Header{"Aaaa": []string{"aaaa aaa aaa aaa"}},
	},
}

func TestHeaderParsing(t *testing.T) {
	for _, tt := range headerTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
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

var statusCodeTests = []struct {
	name    string
	request []byte
	want    string
}{
	{
		name: "MultilineContinuationOnFirstLine",
		request: []byte("GET / HTTP/1.1\r\n" +
			" A: a\r\n" +
			"Host: localhost:8080\r\n" +
			"\r\n"),
		want: statusBadRequest,
	},
	{
		name: "MultilineHeaderName",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"AA\r\n" +
			" AA: aaaa\r\n" +
			"\r\n"),
		want: statusBadRequest,
	},
	{
		name: "MultilineHeaderBeforeColon",
		request: []byte("GET / HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"A\r\n" +
			" : a\r\n" +
			"\r\n"),
		want: statusBadRequest,
	},
}

func TestStatusCode(t *testing.T) {
	for _, tt := range statusCodeTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, nil)
			if err != nil {
				t.Errorf("MakeRequest() got: %v want: nil", err)
			}

			if !bytes.HasPrefix(resp, []byte(tt.want)) {
				got := string(resp[:bytes.IndexByte(resp, '\n')+1])
				t.Errorf("status code got: %q want: %q", got, tt.want)
			}
		})
	}
}
