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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/testing/requesttesting"
	"net/http"
	"net/url"
	"testing"
)

const statusOKReq = "HTTP/1.1 200 OK\r\n"
const statusBadReq = "HTTP/1.1 400 Bad Request\r\n"

// Ensures Query() only returns a map of size one and verifies whether sending a
// request with two values for the same key returns a []string of length 2
// containing the correct values
func TestMultipleQueryParametersSameKey(t *testing.T) {
	const req = "GET /?vegetable=potatO&vegetable=Tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(req), func(req *http.Request) {
		want := url.Values{
			"vegetable": []string{"potatO", "Tomato"},
		}
		got := req.URL.Query()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("req.URL.Query() = %v, want %v, diff (-want +got):\n%s", got, want, diff)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}
}

func TestQueryParametersSameKeyDifferentCasing(t *testing.T) {
	req := []byte("GET /?vegetable=potato&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
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
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}
}

// TestQueryParametersValidUnicode ensures keys and values that contain non-ASCII characters are parsed correctly
func TestQueryParametersValidUnicode(t *testing.T) {
	const value = "ăȚâȘî"
	req := []byte("GET /?vegetable=ăȚâȘî HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		if valueReceived := req.URL.Query()["vegetable"][0]; valueReceived != value {
			t.Errorf("queryParams values: got %s, want %s", valueReceived, value)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}

	key := "ăȚâȘî"
	req = []byte("GET /?" + key + "=vegetable HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		if listLen := len(req.URL.Query()[key]); listLen != 1 {
			t.Errorf("len(queryParamsKey): got %d, want 1 value for %s", listLen, key)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}
}

// Tests whether passing invalid Unicode will result  in a 400 Bad Request response
func TestQueryParametersInvalidUnicodes(t *testing.T) {
	req := []byte("GET /?\x0F=tomato&Vegetable=potato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		t.Fatal("MakeRequest(): Expected handler not to be called for request.")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusBadReq)
	}

}

// Verifies behaviour of query parameter parser when passing malformed keys or
// values, by breaking URL percent-encoding. Percent-encoding is used to encode
// special characters in URL
func TestQueryParametersBreakUrlEncoding(t *testing.T) {
	brokenKeyReq := []byte("GET /?vegetable%=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), brokenKeyReq, func(req *http.Request) {
		if lenQueryParams := len(req.URL.Query()); lenQueryParams != 0 {
			t.Errorf("len(queryParams): got %d, want %d", lenQueryParams, 0)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}

	brokenValueReq := []byte("GET /?vegetable=tomato% HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), brokenValueReq, func(req *http.Request) {
		if lenVeg := len(req.URL.Query()["vegetable"]); lenVeg != 0 {
			t.Errorf("len(queryParams): got %d, want %d", lenVeg, 0)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}

	if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusOKReq)
	}

}

// Test whether both + and %20 are interpreted as space. Having a space in the
// actual request will not be escaped and will result in a 400
func TestQueryParametersSpaceBehaviour(t *testing.T) {
	var tests = []struct {
		name string
		req  []byte
		want string
	}{
		{
			name: "+ as encoding for space",
			req: []byte("GET /?vegetable=potato+pear&Vegetable=tomato HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n"),
			want: "potato pear",
		},
		{
			name: "%20 as encoding for space",
			req: []byte("GET /?vegetable=potato%20pear&Vegetable=tomato HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n"),
			want: "potato pear",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), test.req, func(req *http.Request) {
				if queryParam := req.URL.Query()["vegetable"][0]; queryParam != test.want {
					t.Errorf("URL.Query():want \"potato pear\", got %s", queryParam)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(statusOKReq)) {
				t.Errorf("response status: got %s, want %s", resp, statusOKReq)
			}
		})
	}

	req := []byte("GET /?vegetable=potato pear&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		t.Fatal("Expected handler not to be called with request containing invalid Unicode.")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(statusBadReq)) {
		t.Errorf("response status: got %s, want %s", resp, statusBadReq)
	}
}
