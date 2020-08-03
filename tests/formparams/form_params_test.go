// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package formparams provides tests to inspect the behaviour of form
// parameters parsing in HTTP requests
package formparams

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/internal/requesttesting"
)

const status200OK = "HTTP/1.1 200 OK\r\n"
const status400BadReq = "HTTP/1.1 400 Bad Request\r\n"

func TestSimpleFormParameters(t *testing.T) {
	reqBody := "vegetable=potato&fruit=apple"
	postReq := []byte("POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" +
		"Content-Length: " + strconv.Itoa(len(reqBody)) + "\r\n" +
		"\r\n" +
		reqBody + "\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if err := req.ParseForm(); err != nil {
			t.Errorf("req.ParseForm(): got %v, want nil", err)
		}
		if !cmp.Equal([]string{"potato"}, req.Form["vegetable"]) {
			t.Errorf(`req.Form["vegetable"] = %v, want { "potato" }`, req.Form["vegetable"])
		}
		if !cmp.Equal([]string{"apple"}, req.Form["fruit"]) {
			t.Errorf(`req.Form["fruit"] = %v, want { "apple" }`, req.Form["apple"])
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", string(resp), status200OK)
	}
}

// Test whether passing a POST request with body and without Content-Length
// yields a 400 Bad Request
func TestFormParametersMissingContentLength(t *testing.T) {
	postReq := []byte("POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" +
		"veggie=potato\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		t.Fatal("expected handler not to be called when the request is missing the Content-Length header")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
		t.Errorf("response status: got %s, want %s", string(resp), status400BadReq)
	}
}

// Ensure passing a negative Content-Length or one that overflows integers results in a
// 400 Bad Request
func TestFormParametersBadContentLength(t *testing.T) {
	tests := []struct {
		name string
		req  []byte
	}{
		{
			name: "Negative Content-Length",
			req: []byte("POST / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" +
				"Content-Length: -1\r\n" +
				"\r\n" +
				"veggie=potato\r\n" +
				"\r\n"),
		},
		{
			name: "Integer Overflow Content-Length",
			req: []byte("POST / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" +
				"Content-Length: 10000000000000000000000000\r\n" +
				"\r\n" +
				"veggie=potato\r\n" +
				"\r\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), test.req, func(req *http.Request) {
				t.Error("expected handler not to be called with invalid Content-Length headers")
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
				t.Errorf("response status: got %s, want %s", string(resp), status400BadReq)
			}
		})
	}
}

// Tests behaviour when multiple Content-Length values are passed using three
// testcases:
// a) passing two Content-Length headers containing the same value
// b) passing two Content-Length headers containing different values
// c) passing one Content-Length header containing a list of equal values
// For a) and c), RFC 7320, Section 3.3.2 says that the server can either accept
// the request and use the duplicated value or reject it and for b) it should be
// rejected. Go will handle a) as a valid request and c) as an invalid requet, as the following two tests demonstrate

func TestFormParametersValidDuplicateContentLength(t *testing.T) {
	postReq := []byte("POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" +
		"Content-Length: 13\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"veggie=potato&veggie=a\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		want := []string{"13"}
		got := req.Header["Content-Length"]
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("req.Header: got %v, want %v, diff (-want +got): \n%s", got, want, diff)
		}
		if err := req.ParseForm(); err != nil {
			t.Errorf("req.ParseForm(): got %v, want nil", err)
		}
		if want := []string{"potato"}; !cmp.Equal(want, req.Form["veggie"]) {
			t.Errorf(`Form["veggie"]: got %v, want %v`, req.Form["veggie"], want)
		}

	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}

}

func TestFormParametersInvalidDuplicateContentLength(t *testing.T) {
	tests := []struct {
		name string
		req  []byte
	}{
		{
			name: "Two Content-Length Headers",
			req: []byte("POST / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Type: application/x-www-form-urlencoded; charset=utf-8\r\n" +
				"Content-Length: 13\r\n" +
				"Content-Length: 12\r\n" +
				"\r\n" +
				"veggie=potato\r\n" +
				"\r\n"),
		},
		{
			name: "List of same value in Content-Length Header",
			req: []byte("POST / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Type: application/x-www-form-urlencoded; charset=utf-8\r\n" +
				"Content-Length: 13, 13\r\n" +
				"\r\n" +
				"veggie=potato\r\n" +
				"\r\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), test.req, func(req *http.Request) {
				t.Error("expected handler not to be called")
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
				t.Errorf("response status: got %s, want %s", resp, status400BadReq)
			}
		})
	}
}

// Test whether form parameters with Content-Type:
// application/x-www-form-urlencoded that break percent-encoding will be
// ignored and following valid query parameters will be discarded
func TestFormParametersBreakUrlEncoding(t *testing.T) {
	postReq := []byte("POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/x-www-form-urlencoded\r\n" +
		"Content-Length: 22\r\n" +
		"\r\n" +
		"veggie=%sc&fruit=apple\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if err := req.ParseForm(); err == nil {
			t.Errorf(`req.ParseForm(): got %v, want invalid URL escape`, err)
		}
		var want []string
		got := req.Form["veggie"]
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(`Form["veggie"]: got %v, want %v`, req.Form["vegetable"], nil)
		}
		want = []string{"apple"}
		got = req.Form["fruit"]
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(`Form["fruit"]: got %v, want %v`, req.Form["fruit"], want)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}
}

func TestBasicMultipartForm(t *testing.T) {
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" +
		"Content-Length: " + strconv.Itoa(len(reqBody)) + "\r\n" +
		"\r\n" +
		// CRLF is already in reqBody
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		if err := req.ParseMultipartForm(1024); err != nil {
			t.Fatalf("req.ParseMultipartForm(): want nil, got %v", err)
		}
		if want := []string{"bar"}; !cmp.Equal(want, req.Form["foo"]) {
			t.Errorf("req.Form[foo]: got %v, want %s", req.Form["foo"], want)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}
}

// Test whether a multipart/form-data request passed with no Content-Length
// will be rejected as a bad request by the server and return a 400
func TestMultipartFormNoContentLength(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. The testcase shows the body will be ignored and ParseMultiPartForm will fail
	// with NextPart:EOF.  This is rather inconsistent with
	// application/x-www-form-urlencoded where a  missing Content-Length will
	// return 400 rather than 200.
	t.Skip()
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" +
		"\r\n" +
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("expected handler not to be called")
	})

	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
		t.Errorf("response status: got %s, want %s", resp, status400BadReq)
	}
}

// Test whether passing a multipart/form-data request results in a 400 Bad
// Request server response.
func TestMultipartFormSmallContentLength(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. The test will reach the
	// handler and  result in the error ParseMultipartForm: NextPart: EOF rather
	// than be rejected by the server.
	t.Skip()
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" +
		"Content-Length: 10\r\n" +
		"\r\n" +
		// reqBody ends with CRLF
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("expected handler not to be called")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
		t.Errorf("response status: got %s, want %s", resp, status400BadReq)
	}
}

// Test whether passing a multipart/form-data request with bigger
// Content-Length results in the server blocking, waiting for the entire request
// to be sent.
func TestMultipartFormBigContentLength(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. The handler will be
	// actually called as soon as the first part of the body is received, will
	// return the parsed results and only then block waiting for the rest of the
	// body. This can lead to a Denial of Service attack.
	t.Skip()
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" +
		"Content-Length: 10000000000000000000\r\n" +
		"\r\n" +
		reqBody +
		"\r\n"
	_, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("expected handler not to be called")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
}

// Test behaviour of multipart/form-data when the boundary appears in the
// content as well. According to RFC7578 section 4.1, the request should be
// rejected with a 400 Bad Request.
func TestMultipartFormIncorrectBoundary(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. The request will
	// actually be considered valid by the server and parsed successfully.
	t.Skip()
	reqBody := "--eggplant\r\n" +
		"Content-Disposition: form-data; name=\"eggplant\"\r\n" +
		"\r\n" +
		"eggplant\r\n" +
		"--eggplant--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"eggplant\"\r\n" +
		"Content-Length: " + strconv.Itoa(len(reqBody)) + "\r\n" +
		"\r\n" +
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("expected handler not to be called")
	})

	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
		t.Errorf("response status: got %s, want %s", resp, status400BadReq)
	}
}

// TODO(mihalimara22@): Since req.Form["x"] and req.FormValue("x") return different
// types, this aspect should be investigated more in future tests
