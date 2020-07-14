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

package requestparsing

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-safeweb/testing/requesttesting"

	"github.com/google/go-cmp/cmp"
)

func TestContentLengthTransferEncoding(t *testing.T) {
	type testWant struct {
		headers          map[string][]string
		contentLength    int64
		transferEncoding []string
		body             string
	}

	var tests = []struct {
		name    string
		request []byte
		want    testWant
	}{
		{
			name: "ContentLength",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 5\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{"Content-Length": []string{"5"}},
				contentLength:    5,
				transferEncoding: nil,
				body:             "ABCDE",
			},
		},
		{
			name: "ContentButNoContentLength",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    0,
				transferEncoding: nil,
				body:             "",
			},
		},
		{
			name: "TransferEncodingChunked",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "ContentLengthAndTransferEncoding",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "MultipleTransferEncodingChunkedFirst",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"Transfer-Encoding: asdf\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "TransferEncodingIdentity",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{"Content-Length": []string{"11"}},
				contentLength:    11,
				transferEncoding: nil,
				body:             "5\r\nABCDE\r\n0",
			},
		},
		{
			name: "TransferEncodingListIdentityFirst",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: identity, xyz, asdf\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{"Content-Length": []string{"11"}},
				contentLength:    11,
				transferEncoding: nil,
				body:             "5\r\nABCDE\r\n0",
			},
		},
		{
			name: "TransferEncodingListChunkedIdentity",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked, identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "TransferEncodingCasingOrdering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"transfer-Encoding: identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "TransferEncodingCasingOrdering2",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"transfer-Encoding: chunked\r\n" +
				"Transfer-Encoding: identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want.headers, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.want.transferEncoding, r.TransferEncoding); diff != "" {
					t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
				}

				if r.ContentLength != tt.want.contentLength {
					t.Errorf("r.ContentLength got: %v want: %v", r.ContentLength, tt.want.contentLength)
				}

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("ioutil.ReadAll(r.Body) got err: %v", err)
				}

				if got := string(body); got != tt.want.body {
					t.Errorf("r.Body got: %q want: %q", got, tt.want.body)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got, want := extractStatus(resp), statusOK; got != want {
				t.Errorf("status code got: %q want: %q", got, want)
			}
		})
	}
}

func TestContentLengthTransferEncodingStatusMessages(t *testing.T) {
	var tests = []struct {
		name    string
		request []byte
		want    string
	}{
		{
			name: "MultipleContentLength",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 5\r\n" +
				"Content-Length: 14\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n"),
			want: statusBadRequest,
		},
		{
			name: "MultipleTransferEncodingChunkedSecond",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: asdf\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: statusNotImplemented,
		},
		{
			name: "TransferEncodingListChunkedFirst",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked, xyz, asdf\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: statusNotImplemented,
		},
		{
			name: "TransferEncodingListChunkedChunked",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked, chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: statusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, nil)
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got := extractStatus(resp); got != tt.want {
				t.Errorf("status code got: %q want: %q", got, tt.want)
			}
		})
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
			t.Fatalf("ioutil.ReadAll(r.Body) got: %v want: http chunk length too large", err)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest() got err: %v", err)
	}
}
