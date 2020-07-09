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
	"context"
	"go-safeweb/testing/requesttesting"
	"log"
	"net/http"
	"strconv"
	"testing"
)

func TestSimpleFormParameters(t *testing.T) {
	reqBody := "Veg=Potato&Fruit=Apple"
	reqBodyLen := len(reqBody)
	postReq := []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length: " + strconv.Itoa(reqBodyLen) + "\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		req.ParseForm()
		if req.Form["Veg"][0] != "Potato" && req.Form["Fruit"][0] != "Apple" {
			t.Errorf("req.Form: want Potato and Apple, got %s and %s", req.Form["Veg"][0], req.Form["Fruit"][0])
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}
}

// Test whether passing a POST request with body and without Content-Length
// yields a 400 Bad Request
func TestFormParametersMissingContentLength(t *testing.T) {
	reqBody := "veggie=potato"
	postReq := []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + reqBody + "\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		t.Error("MakeRequest(): Expected handler not to be called.")
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		t.Errorf("response: want %s, got %s", statusBadReq, string(resp))
	}
}

// Ensure passing a negative Content-Length or one that overflows results in a
// 400 Bad Request
func TestFormParametersBadContentLength(t *testing.T) {
	reqBody := "veggie=potato"
	postReq := []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length: -1" + reqBody + "\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		t.Error("MakeRequest(): Expected handler not to be called.")
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		t.Errorf("response: want %s, got %s", statusBadReq, string(resp))
	}

	postReq = []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length: 10000000000000000000000000" + reqBody + "\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		t.Error("MakeRequest(): Expected handler not to be called.")
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		t.Errorf("response: want %s, got %s", statusBadReq, string(resp))
	}
}

// Tests behaviour when multiple Content-Length are passed. If they are
// duplicated and of equal
// length, the request will succeed, otherwise it fails. This conforms to RFC
// 7320, Section 3.3.2 which says that the server can either replace duplicate
// Content-Length headers with one value or reject the request. However, for a
// list of identical values the request is rejected, which is rather inconsistent.
func TestFormParametersDuplicateContentLength(t *testing.T) {
	t.Skip("Content-Length test skipped.")
	reqBody := "veggie=potato"
	postReq := []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=utf-8\r\n" + "Content-Length: 13\r\n" + "Content-Length: 13\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if contentLen := req.Header["Content-Length"][0]; contentLen != "13" {
			t.Errorf("Expected Content-Length to be 13 but got %s", contentLen)
			if val := req.FormValue("veggie"); val != "potato" {
				t.Errorf("req.FormValue: want potato, got %s", val)
			}
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

	reqBody = "veggie=potato"
	postReq = []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=utf-8\r\n" + "Content-Length: 13, 13\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if contentLen := req.Header["Content-Length"][0]; contentLen != "13" {
			t.Errorf("Expected Content-Length to be 13 but got %s", contentLen)
			if val := req.FormValue("veggie"); val != "potato" {
				t.Errorf("req.FormValue: want potato, got %s", val)
			}
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

	postReq = []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length: 13\r\n" + "Content-Length: 12\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if contentLen := req.Header["Content-Length"][0]; contentLen != "13" {
			t.Error("MakeRequest(): request should have been rejected when list of values is given to Content-Length, rather inconsistent if you ask me.")
		}
		req.ParseForm()

	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		t.Errorf("response: want %s, got %s", statusBadReq, string(resp))
	}
}

// Test whether form parameters with Content-Type:
// application/x-www-form-urlencoded that break percent-encoding will be
// ignored
func TestFormParametersBreakUrlEncoding(t *testing.T) {

	reqBody := "veggie=%sc"
	reqBodyLen := len(reqBody)
	postReq := []byte("POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded\r\n" + "Content-Length:" + strconv.Itoa(reqBodyLen) + "\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), postReq, func(req *http.Request) {
		if val := req.Form["veggie"]; len(req.Form["veggie"]) != 0 {
			t.Errorf("req.Form[\"veggie\"]: got %s, want %s.", val, "")
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

}

func TestBasicMultipartForm(t *testing.T) {
	reqBody := "--123\r\n" + "Content-Disposition: form-data; name=\"foo\"\r\n" + "\r\n" + "bar\r\n" + "--123--\r\n"
	reqBodyLen := len(reqBody)
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: " + strconv.Itoa(reqBodyLen) + "\r\n" + "\r\n" + reqBody + "\r\n\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {

		e := req.ParseMultipartForm(1024)
		if e != nil {
			log.Printf("ParseMultipartForm: want nil, got %v", e)
		}

		if formVal := req.Form["foo"][0]; formVal != "bar" {
			log.Printf("FormValue: want bar, got %s", formVal)
		}
	})
	if err != nil {
		log.Printf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		log.Printf("response: want %s, got %s", statusOK, string(resp))
	}
}

// Test behaviour of multipart/form-data request passed with no Content-Length.
// In this case, the body will be ignored and ParseMultiPartForm()å parsing will fail
// with NextPart:EOF.  This is rather inconsistent with
// application/x-www-form-urlencoded where a  missing Content-Length will
// return 400 rather than 200.
func TestMultipartFormNoContentLength(t *testing.T) {
	reqBody := "--123\r\n" + "Content-Disposition: form-data; name=\"foo\"\r\n" + "\r\n" + "bar\r\n" + "--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "\r\n" + "\r\n" + reqBody + "\r\n\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})

	if err != nil {
		log.Printf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		log.Printf("response: want %s, got %s", statusBadReq, string(resp))
	}
}

// Test behaviour of multipart/form-data request passed with smaller
// Content-Length than the actual body. It will trigger a parsing error
// ParseMultipartForm: NextPart: EOF, but a
// more consistent behaviour should be to return a 400 rather than 200
func TestMultipartFormSmallContentLength(t *testing.T) {
	reqBody := "--123\r\n" + "Content-Disposition: form-data; name=\"foo\"\r\n" + "\r\n" + "bar\r\n" + "--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: 10" + "\r\n" + "\r\n" + reqBody + "\r\n\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})
	if err != nil {
		log.Printf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		log.Printf("response: want %s, got %s", statusBadReq, string(resp))
	}
}

// Test behaviour of multipart/form-data request passed with bigger
// Content-Length than the actual body. It will parse successfully and then
// block waiting for the rest of the body. Calling the handler before the entire
// body has been received might not be the best choice.
func TestMultipartFormBigContentLength(t *testing.T) {
	reqBody := "--123\r\n" + "Content-Disposition: form-data; name=\"foo\"\r\n" + "\r\n" + "bar\r\n" + "--123--\r\n"
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: 10000000000000000000" + "\r\n" + "\r\n" + reqBody + "\r\n\r\n"
	_, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})
	if err != nil {
		log.Printf("MakeRequest(): got err %v, want nil", err)
	}
}

// Test behaviour of multipart/form-data when the boundary appears in the
// content as well. According to RFC7578 section 4.1 (as far as I understand),
// this should not be allowed but the request succeeds.
func TestMultipartFormIncorrectBoundary(t *testing.T) {
	reqBody := "--eggplant\r\n" + "Content-Disposition: form-data; name=\"eggplant\"\r\n" + "\r\n" + "eggplant\r\n" + "--eggplant--\r\n"
	reqBodyLen := len(reqBody)
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
		"Content-Type: multipart/form-data; boundary=\"eggplant\"\r\n" + "Content-Length: " + strconv.Itoa(reqBodyLen) + "\r\n" + "\r\n" + reqBody + "\r\n\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(postReq), func(req *http.Request) {
		t.Error("MakeRequest(): expected handler not to be called")
	})

	if err != nil {
		log.Printf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		log.Printf("response: want %s, got %s", statusBadReq, string(resp))
	}
}
