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

// Package queryparams provides tests to inspect the behaviour of query
// parameters parsing in HTTP requests
package queryparams

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/internal/requesttesting"
)

const status200OK = "HTTP/1.1 200 OK\r\n"
const status400BadReq = "HTTP/1.1 400 Bad Request\r\n"

// Ensures Query() only returns a map of size one and verifies whether sending a
// request with two values for the same key returns a []string of length 2
// containing the correct values
func TestMultipleQueryParametersSameKey(t *testing.T) {
	const req = "GET /?vegetable=potatO&vegetable=Tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n"
	resp, err := requesttesting.MakeRequest(context.Background(), []byte(req), func(req *http.Request) {
		want := url.Values{"vegetable": []string{"potatO", "Tomato"}}
		got := req.URL.Query()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("req.URL.Query(): got %v, want %v, diff (-want +got):\n%s", got, want, diff)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}
}

func TestQueryParametersSameKeyDifferentCasing(t *testing.T) {
	req := []byte("GET /?vegetable=potato&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		want := url.Values{
			"vegetable": []string{"potato"},
			"Vegetable": []string{"tomato"},
		}
		got := req.URL.Query()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("req.URL.Query(): got %v, want %v, diff (-want +got): \n%s", got, want, diff)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}
}

// TestQueryParametersValidUnicode ensures keys and values that contain non-ASCII characters are parsed correctly
func TestQueryParametersValidUnicode(t *testing.T) {
	req := []byte("GET /?vegetable=ăȚâȘî HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		want := url.Values{"vegetable": []string{"ăȚâȘî"}}
		got := req.URL.Query()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("req.URL.Query(): got %v, want %v, diff (-want +got):\n%s", got, want, diff)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}

	req = []byte("GET /?ăȚâȘî=vegetable HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err = requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		want := url.Values{"ăȚâȘî": []string{"vegetable"}}
		got := req.URL.Query()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("req.URL.Query(): got %v, want %v, diff (-want +got):\n%s", got, want, diff)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status200OK)) {
		t.Errorf("response status: got %s, want %s", resp, status200OK)
	}
}

// Tests whether passing invalid Unicode in both key and value will result in a 400 Bad Request response
func TestQueryParametersInvalidUnicodes(t *testing.T) {

	tests := []struct {
		name string
		req  []byte
	}{
		{
			name: "Invalid Key",
			req: []byte("GET /?\x0F=tomato&Vegetable=potato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
				"\r\n"),
		},
		{
			name: "Invalid value",
			req: []byte("GET /?vegetable=\x0F&Vegetable=potato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
				"\r\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), test.req, func(req *http.Request) {
				t.Error("expected handler not to be called.")
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
				t.Errorf("response status: got %s, want %s", resp, status400BadReq)
			}
		})
	}

}

// Verifies behaviour of query parameter parser when passing malformed keys or
// values, by breaking URL percent-encoding. Percent-encoding is used to encode
// special characters in URL
func TestQueryParametersBreakUrlEncoding(t *testing.T) {
	tests := []struct {
		name string
		req  []byte
	}{
		{
			name: "Broken Key",
			req: []byte("GET /?vegetable%=tomato&Vegetable=potato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
				"\r\n"),
		},
		{
			name: "Broken value",
			req: []byte("GET /?vegetable=tomato%&Vegetable=potato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" +
				"\r\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), test.req, func(req *http.Request) {
				var want []string
				got := req.URL.Query()["vegetable"]
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf(`URL.Query()["vegetable"]:, got %v, want %v`, req.URL.Query()["vegetable"], want)
				}
				want = []string{"potato"}
				got = req.URL.Query()["Vegetable"]
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf(`URL.Query()["Vegetable"]:, got %v, want %v`, req.URL.Query()["vegetable"], want)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(status200OK)) {
				t.Errorf("response status: got %s, want %s", resp, status200OK)
			}
		})
	}

}

// Test whether both + and %20 are interpreted as space. Having a space in the
// actual request will not be escaped and will result in a 400
func TestQueryParametersSpaceBehaviour(t *testing.T) {
	tests := []struct {
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
				if want := []string{"potato pear"}; !cmp.Equal(want, req.URL.Query()["vegetable"]) {
					t.Errorf("URL.Query():, got %v, want %v", req.URL.Query()["vegetable"], want)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest(): got err %v, want nil", err)
			}
			if !bytes.HasPrefix(resp, []byte(status200OK)) {
				t.Errorf("response status: got %s, want %s", resp, status200OK)
			}
		})
	}

	req := []byte("GET /?vegetable=potato pear&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n")
	resp, err := requesttesting.MakeRequest(context.Background(), req, func(req *http.Request) {
		t.Fatal("expected handler not to be called with request containing invalid Unicode.")
	})
	if err != nil {
		t.Fatalf("MakeRequest(): got err %v, want nil", err)
	}
	if !bytes.HasPrefix(resp, []byte(status400BadReq)) {
		t.Errorf("response status: got %s, want %s", resp, status400BadReq)
	}
}
