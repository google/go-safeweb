// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/google/go-safeweb/testing/requesttesting"
	"net/http"
	"testing"
)

const statusOK = "HTTP/1.1 200 OK\r\n"
const statusBadReq = "HTTP/1.1 400 Bad Request\r\n"

// isOkReq is a helper function that checks whether the server response is 200
func isOkReq(resp []byte) bool {
	if bytes.HasPrefix(resp, []byte(statusOK)) {
		return true
	}
	return false
}

// isBadReq is a helper function that checks whether the server response is 400
func isBadReq(resp []byte) bool {
	if bytes.HasPrefix(resp, []byte(statusBadReq)) {
		return true
	}
	return false
}

// Ensures Query() only returns a map of size one and verifies whether sending a
// request with two values for the same key returns a []string of length 2
// containing the correct values
func TestMultipleQueryParametersSameKey(t *testing.T) {
	var (
		valueOne = "potatO"
		valueTwo = "Tomato"
		req      = []byte("GET /?vegetable=" + valueOne + "&vegetable=" + valueTwo + " HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n")
	)
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		queryParams := req.URL.Query()
		if len(queryParams) != 1 {
			t.Errorf("len(queryParams): got %d, want %d", len(queryParams), 1)
		}

		vegetableParamValues := queryParams["vegetable"]
		if len(vegetableParamValues) != 2 {
			t.Errorf("len(vegetableQueryParams): got %d, want %d", len(vegetableParamValues), 2)
		}
		if vegetableParamValues[0] != valueOne || vegetableParamValues[1] != valueTwo {
			t.Errorf("queryParams values: expected "+valueOne+" and "+valueTwo+", got %s and %s", vegetableParamValues[0], vegetableParamValues[1])
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

}

func TestQueryParametersSameKeyDifferentCasing(t *testing.T) {

	req := []byte("GET /?vegetable=potato&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		queryParams := req.URL.Query()

		if len(queryParams) != 2 {
			t.Errorf("len(queryParams): got %d, want %d", len(queryParams), 2)
		}

		if len(queryParams["vegetable"]) != 1 || len(queryParams["Vegetable"]) != 1 {
			t.Errorf("Expected one value for query parameter vegetables and one for Vegetables, got %d and %d", len(queryParams["vegetable"]), len(queryParams["Vegetable"]))
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}
}

// TestQueryParametersValidUnicode ensures keys and values that contain non-ASCII characters are parsed correctly
func TestQueryParametersValidUnicode(t *testing.T) {
	value := "ăȚâȘî"
	req := []byte("GET /?vegetable=" + value + " HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		if valueReceived := req.URL.Query()["vegetable"][0]; valueReceived != value {
			t.Errorf("queryParams values: got %s, want %s", valueReceived, value)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

	key := "ăȚâȘî"
	req = []byte("GET /?" + key + "=vegetable HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		if listLen := len(req.URL.Query()[key]); listLen != 1 {
			t.Errorf("len(queryParamsKey): got %d, want 1 value for %s", listLen, key)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

}

// Tests whether passing invalid Unicode will result  in a 400 Bad Request response
func TestQueryParametersInvalidUnicodes(t *testing.T) {
	key := "\x0F"
	req := []byte("GET /?" + key + "=tomato&Vegetable=potato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		t.Error("MakeRequest(): Expected handler not to be called.")
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		t.Errorf("response: want %s, got %s", statusBadReq, string(resp))
	}

}

// Verifies behaviour of query parameter parser when passing malformed keys or
// values, by breaking URL percent-encoding. Percent-encoding is used to encode
// special characters in URL
func TestQueryParametersBreakUrlEncoding(t *testing.T) {
	brokenKeyReq := []byte("GET /?vegetable%=tomato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), brokenKeyReq, func(req *http.Request) {
		if lenQueryParams := len(req.URL.Query()); lenQueryParams != 0 {
			t.Errorf("len(queryParams): got %d, want %d", lenQueryParams, 0)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

	brokenValueReq := []byte("GET /?vegetable=tomato% HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), brokenValueReq, func(req *http.Request) {
		if lenVeg := len(req.URL.Query()["vegetable"]); lenVeg != 0 {
			t.Errorf("len(queryParams): got %d, want %d", lenVeg, 0)
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}

	if !isOkReq(resp) {
		t.Errorf("response: want %s, got %s", statusOK, string(resp))
	}

}

// Test whether both + and %20 are interpreted as space. Having a space in the
// actual request will not be escaped and will result in a 400
func TestQueryParametersSpaceBehaviour(t *testing.T) {
	var (
		value1 string
		value2 string
	)
	req := []byte("GET /?vegetable=potato pear&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		if req != nil {
			t.Error("Expected the server not to receive a request containing invalid Unicode.")
		}
	})
	if err != nil {
		t.Errorf("MakeRequest(): got err %v, want nil", err)
	}
	if !isBadReq(resp) {
		t.Errorf("response: want %s, got %s", statusBadReq, string(resp))
	}

	req = []byte("GET /?vegetable=potato+pear&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		value1 = req.URL.Query()["vegetable"][0]
	})

	req = []byte("GET /?vegetable=potato%20pear&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		value2 = req.URL.Query()["vegetable"][0]
	})

	if value1 != value2 {
		t.Errorf("URL.Query():want potato and potato but got %s and %s", value1, value2)
	}
}
