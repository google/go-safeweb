package requestparsing

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-safeweb/testing/requesttesting"

	"github.com/google/go-cmp/cmp"
)

func TestContentLength(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 5\r\n" +
		"\r\n" +
		"ABCDE\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		want := map[string][]string{"Content-Length": []string{"5"}}
		if diff := cmp.Diff(want, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != 5 {
			t.Errorf("r.ContentLength got: %v want: 5", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "ABCDE" {
			t.Errorf(`r.Body got: %q want: "ABCDE"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestMultipleContentLength(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 5\r\n" +
		"Content-Length: 14\r\n" +
		"\r\n" +
		"ABCDE\r\n" +
		"\r\n" +
		"ABCDE\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, nil)
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusBadRequest)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusBadRequest)
	}
}

func TestContentButNoContentLength(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n" +
		"ABCDE\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != 0 {
			t.Errorf("r.ContentLength got: %v want: 0", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "" {
			t.Errorf(`r.Body got: %q want: ""`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestTransferEncodingChunked(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != -1 {
			t.Errorf("r.ContentLength got: %v want: -1", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "ABCDE" {
			t.Errorf(`r.Body got: %q want: "ABCDE"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestContentLengthAndTransferEncoding(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != -1 {
			t.Errorf("r.ContentLength got: %v want: -1", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "ABCDE" {
			t.Errorf(`r.Body got: %q want: "ABCDE"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestMultipleTransferEncodingChunkedFirst(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"Transfer-Encoding: asdf\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != -1 {
			t.Errorf("r.ContentLength got: %v want: -1", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "ABCDE" {
			t.Errorf(`r.Body got: %q want: "ABCDE"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestMultipleTransferEncodingChunkedSecond(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: asdf\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, nil)
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusNotImplemented)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusNotImplemented)
	}
}

func TestTransferEncodingIdentity(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: identity\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		want := map[string][]string{"Content-Length": []string{"11"}}
		if diff := cmp.Diff(want, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string(nil), r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != 11 {
			t.Errorf("r.ContentLength got: %v want: 11", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "5\r\nABCDE\r\n0" {
			t.Errorf(`r.Body got: %q want: "5\r\nABCDE\r\n0"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestTransferEncodingListIdentityFirst(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: identity, xyz, asdf\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		want := map[string][]string{"Content-Length": []string{"11"}}
		if diff := cmp.Diff(want, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string(nil), r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != 11 {
			t.Errorf("r.ContentLength got: %v want: 11", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "5\r\nABCDE\r\n0" {
			t.Errorf(`r.Body got: %q want: "5\r\nABCDE\r\n0"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestTransferEncodingListChunkedFirst(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: chunked, xyz, asdf\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, nil)
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusNotImplemented)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusNotImplemented)
	}
}

func TestTransferEncodingListChunkedIdentity(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: chunked, identity\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		want := map[string][]string{}
		if diff := cmp.Diff(want, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != -1 {
			t.Errorf("r.ContentLength got: %v want: -1", r.ContentLength)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) got err: %v", err)
		}

		if got := string(body); got != "ABCDE" {
			t.Errorf(`r.Body got: %q want: "ABCDE"`, got)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOK)) {
		got := string(resp[:bytes.IndexByte(resp, '\n')+1])
		t.Errorf("status code got: %q want: %q", got, statusOK)
	}
}

func TestTransferEncodingChunkSizeLength(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host:localhost\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"000000000000000000007\r\n" +
		"\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	_, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != -1 {
			t.Errorf("r.ContentLength got: %v want: -1", r.ContentLength)
		}

		_, err := ioutil.ReadAll(r.Body)
		if err == nil {
			t.Errorf("ioutil.ReadAll(r.Body) got: %v want: http chunk length too large", err)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest() got err: %v", err)
	}
}
