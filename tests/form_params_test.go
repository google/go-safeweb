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

package tests

import (
	"bytes"
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/testing/requesttesting"
	"net/http"
	"strconv"
	"testing"
)

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
		req.ParseForm()
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
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", string(resp), statusOKReq)
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
		t.Error("MakeRequest(): Expected handler not to be called.")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", string(resp), statusBadReq)
	}
}

// Ensure passing a negative Content-Length or one that overflows integers results in a
// 400 Bad Request
func TestFormParametersBadContentLength(t *testing.T) {
	var tests = []struct {
		name string
		req  []byte
	}{
		{
			name: "Negative Content-Length",
			req: []byte("POST / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" +
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
				t.Error("MakeRequest(): expected handler not to be called.")
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
				t.Errorf("response status: got %s, want %s", string(resp), statusBadReq)
			}
		})
	}
}

// Tests behaviour when multiple Content-Length values are passed. Go's
// behaviour is the following:
// a) if we pass two Content-Length headers containing the same value, the // request
// is deemed valid
// b) if we pass two Content-Length headers containing different values, the
// server will respond with a 400 Bad Request
// c) if we pass one Content-Length header containing a list of equal values,
// the server will respond with a 400 Bad Request
// For a) and c), RFC 7320, Section 3.3.2 says that the server can either accept
// the request and use the duplicated value or reject it. However, a more
// consistent behaviour would be to handle a) and c) similarly.
func TestFormParametersDuplicateContentLength(t *testing.T) {
	type want struct {
		contentLen    string
		queryParamVal string
	}
	var tests = []struct {
		name string
		req  []byte
		want want
	}{
		{
			name: "Two Content-Length Headers",
			req: []byte("POST / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Type: application/x-www-form-urlencoded; charset=utf-8\r\n" +
				"Content-Length: 13\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"veggie=potato\r\n" +
				"\r\n"),
			want: want{
				contentLen:    "13",
				queryParamVal: "potato",
			},
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
			want: want{
				contentLen:    "13",
				queryParamVal: "potato",
			},
		},
	}

	for _, test := range tests {
		// Skipping to ensure GitHub checks are not failing. This test is made
		// to fail intentionally to pinpoint behaviour we regard as inconsistent
		t.Skip()
		t.Run(test.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), test.req, func(req *http.Request) {
				if contentLen := req.Header["Content-Length"][0]; contentLen != test.want.contentLen {
					t.Errorf(`req.Header["Content-Length"]: got %v, want %v`, contentLen, test.want.contentLen)
				}
				if val := req.FormValue("veggie"); val != test.want.queryParamVal {
					t.Errorf("req.FormValue: want potato, got %s", val)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !!bytes.HasPrefix(resp, []byte(statusOKReq)) {
				t.Errorf("response status: got %s, want %s", resp, statusOKReq)
			}
		})
	}

	postReq := []byte("POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length: 13\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		"veggie=potato\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if contentLen := req.Header["Content-Length"][0]; contentLen != "13" {
			t.Error("MakeRequest(): request should have been rejected when list of values is given to Content-Length, rather inconsistent if you ask me.")
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusBadReq)
	}
}

// Test whether form parameters with Content-Type:
// application/x-www-form-urlencoded that break percent-encoding will be
// ignored
func TestFormParametersBreakUrlEncoding(t *testing.T) {
	postReq := []byte("POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/x-www-form-urlencoded\r\n" +
		"Content-Length: 10\r\n" +
		"\r\n" +
		"veggie=%sc\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if val := req.Form["veggie"]; len(req.Form["veggie"]) != 0 {
			t.Errorf("req.Form[\"veggie\"]: got %s, want %s.", val, "")
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}
}

func TestBasicMultipartForm(t *testing.T) {
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: " + strconv.Itoa(len(reqBody)) + "\r\n" +
		"\r\n" +
		reqBody + "\r\n" +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {

		e := req.ParseMultipartForm(1024)
		if e != nil {
			t.Fatalf("ParseMultipartForm: want nil, got %v", e)
		}

		if formVal := req.Form["foo"][0]; formVal != "bar" {
			t.Errorf("FormValue: got %s, want bar", formVal)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}
}

// Test behaviour of multipart/form-data request passed with no Content-Length.
// In this case, the body will be ignored and ParseMultiPartForm()å parsing will fail
// with NextPart:EOF.  This is rather inconsistent with
// application/x-www-form-urlencoded where a  missing Content-Length will
// return 400 rather than 200.
func TestMultipartFormNoContentLength(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. This test is made
	// to fail intentionally to pinpoint behaviour we regard as inconsistent
	t.Skip()
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "\r\n" +
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})

	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusBadReq)
	}
}

// Test behaviour of multipart/form-data request passed with smaller
// Content-Length than the actual body. It will trigger a parsing error
// ParseMultipartForm: NextPart: EOF, but a
// more consistent behaviour should be to return a 400 rather than 200
func TestMultipartFormSmallContentLength(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. This test is made
	// to fail intentionally to pinpoint behaviour we regard as inconsistent
	t.Skip()
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: 10\r\n" +
		"\r\n" +
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusBadReq)
	}
}

// Test behaviour of multipart/form-data request passed with bigger
// Content-Length than the actual body. It will parse successfully and then
// block waiting for the rest of the body. Calling the handler before the entire
// body has been received might not be the best choice.
func TestMultipartFormBigContentLength(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. This test is made
	// to fail intentionally to pinpoint behaviour we regard as inconsistent
	t.Skip()
	reqBody := "--123\r\n" +
		"Content-Disposition: form-data; name=\"foo\"\r\n" +
		"\r\n" +
		"bar\r\n" +
		"--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: 10000000000000000000\r\n" +
		"\r\n" +
		reqBody + "\r\n"
	_, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
}

// Test behaviour of multipart/form-data when the boundary appears in the
// content as well. According to RFC7578 section 4.1 (as far as I understand),
// this should not be allowed but the request succeeds.
func TestMultipartFormIncorrectBoundary(t *testing.T) {
	// Skipping to ensure GitHub checks are not failing. This test is made
	// to fail intentionally to pinpoint behaviour we regard as inconsistent
	t.Skip()
	reqBody := "--eggplant\r\n" +
		"Content-Disposition: form-data; name=\"eggplant\"\r\n" +
		"\r\n" +
		"eggplant\r\n" +
		"--eggplant--\r\n"
	postReq := "POST / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"eggplant\"\r\n" + "Content-Length: " + strconv.Itoa(len(reqBody)) + "\r\n" +
		"\r\n" +
		reqBody +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})

	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusBadReq)
	}
}
