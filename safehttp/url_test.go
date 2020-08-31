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

package safehttp

import (
	"net/url"
	"testing"
)

func TestURLToString(t *testing.T) {
	want := "http://www.example.com/asdf?fruit=apple"

	netURL, err := url.Parse(want)
	if err != nil {
		t.Fatalf(`url.Parse("http://www.example.com/asdf?fruit=apple") got: %v want: nil`, err)
	}

	u := URL{url: netURL}
	if got := u.String(); got != want {
		t.Errorf("u.String() got: %v want: %v", got, want)
	}
}

func TestURLHost(t *testing.T) {
	var test = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Just host",
			url:  "http://www.example.com/asdf?fruit=apple",
			want: "www.example.com",
		},
		{
			name: "Host and port",
			url:  "http://www.example.com:1337/asdf?fruit=apple",
			want: "www.example.com:1337",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			netURL, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse(tt.url) got: %v want: nil", err)
			}

			u := URL{url: netURL}
			if got := u.Host(); got != tt.want {
				t.Errorf("u.Host() got: %v want: %v", got, tt.want)
			}
		})
	}
}

func TestURLHostname(t *testing.T) {
	var test = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Just host",
			url:  "http://www.example.com/asdf?fruit=apple",
			want: "www.example.com",
		},
		{
			name: "Host and port",
			url:  "http://www.example.com:1337/asdf?fruit=apple",
			want: "www.example.com",
		},
		{
			name: "Ipv6 and port",
			url:  "http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:17000/asdf?fruit=apple",
			want: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			netURL, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse(tt.url) got: %v want: nil", err)
			}

			u := URL{url: netURL}
			if got := u.Hostname(); got != tt.want {
				t.Errorf("u.Hostname() got: %v want: %v", got, tt.want)
			}
		})
	}
}

func TestURLPort(t *testing.T) {
	var test = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Just host",
			url:  "http://www.example.com/asdf?fruit=apple",
			want: "",
		},
		{
			name: "Host and port",
			url:  "http://www.example.com:1337/asdf?fruit=apple",
			want: "1337",
		},
		{
			name: "HTTPS",
			url:  "https://www.example.com/asdf?fruit=apple",
			want: "",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			netURL, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse(tt.url) got: %v want: nil", err)
			}

			u := URL{url: netURL}
			if got := u.Port(); got != tt.want {
				t.Errorf("u.Port() got: %v want: %v", got, tt.want)
			}
		})
	}
}

func TestURLPath(t *testing.T) {
	netURL, err := url.Parse("http://www.example.com/asdf?fruit=apple")
	if err != nil {
		t.Fatalf(`url.Parse("http://www.example.com/asdf?fruit=apple") got: %v want: nil`, err)
	}

	u := URL{url: netURL}
	if got, want := u.Path(), "/asdf"; got != want {
		t.Errorf("u.Path() got: %v want: %v", got, want)
	}
}

func TestURLQuery(t *testing.T) {
	var test = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Normal",
			url:  "http://www.example.com/asdf?fruit=apple",
			want: "apple",
		},
		{
			name: "Empty",
			url:  "http://www.example.com/asdf",
			want: "",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			netURL, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse(tt.url) got: %v want: nil", err)
			}

			u := URL{url: netURL}
			f, err := u.Query()
			if err != nil {
				t.Errorf("u.Query() got: %v want: nil", err)
			}

			if got := f.String("fruit", ""); got != tt.want {
				t.Errorf(`f.String("fruit", "") got: %q want: %q`, got, tt.want)
			}
		})
	}
}

func TestURLInvalidQuery(t *testing.T) {
	netURL, err := url.Parse("http://www.example.com/asdf?%xx=abc")
	if err != nil {
		t.Fatalf(`url.Parse("http://www.example.com/asdf?%%xx=abc") got: %v want: nil`, err)
	}

	u := URL{url: netURL}
	if _, err = u.Query(); err == nil {
		t.Error("u.Query() got: nil want: error")
	}
}

func TestURLParse(t *testing.T) {
	var tests = []struct {
		name, url string
	}{
		{
			name: "Absolute url",
			url:  "http://www.example.com/path?foo=bar#tar",
		},
		{
			name: "Relative url",
			url:  "example.com/path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, err := Parse(test.url)
			if url == nil {
				t.Error("Parse(test.url): got nil, want URL")
			}
			if err != nil {
				t.Errorf("Parse(test.url): got %v, want nil", err)
			}
			if got := url.String(); got != test.url {
				t.Errorf("url.String(): got %v, want %v", got, test.url)
			}
		})
	}
}

func TestURLInvalidParse(t *testing.T) {
	url, err := Parse("http://www.example.com/path%%%x=0")
	if url != nil {
		t.Errorf(`Parse: got %v, want nil`, url)
	}
	if err == nil {
		t.Errorf(`Parse: got nil, want error`)
	}
}
