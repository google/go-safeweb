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

import "testing"

func TestCookie(t *testing.T) {
	tests := []struct {
		name   string
		cookie *Cookie
		want   string
	}{
		{
			name:   "Default",
			cookie: NewCookie("foo", "bar"),
			want:   "foo=bar; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "SameSite Lax",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.SameSite(SameSiteNoneMode)
				c.SameSite(SameSiteLaxMode)
				return c
			}(),
			want: "foo=bar; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "SameSite strict",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.SameSite(SameSiteStrictMode)
				return c
			}(),
			want: "foo=bar; HttpOnly; Secure; SameSite=Strict",
		},
		{
			name: "SameSite none",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.SameSite(SameSiteNoneMode)
				return c
			}(),
			want: "foo=bar; HttpOnly; Secure; SameSite=None",
		},
		{
			name: "Maxage positive",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.SetMaxAge(10)
				return c
			}(),
			want: "foo=bar; Max-Age=10; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "Maxage negative",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.SetMaxAge(-1)
				return c
			}(),
			want: "foo=bar; Max-Age=0; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "Path",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.Path("/asdf")
				return c
			}(),
			want: "foo=bar; Path=/asdf; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "Domain",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.Domain("example.com")
				return c
			}(),
			want: "foo=bar; Domain=example.com; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "Not Secure",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.DisableSecure()
				return c
			}(),
			want: "foo=bar; HttpOnly; SameSite=Lax",
		},
		{
			name: "Not HttpOnly",
			cookie: func() *Cookie {
				c := NewCookie("foo", "bar")
				c.DisableHTTPOnly()
				return c
			}(),
			want: "foo=bar; Secure; SameSite=Lax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cookie.String(); got != tt.want {
				t.Errorf("tt.cookie.String() got: %q want: %q", got, tt.want)
			}
		})
	}
}

func TestCookieName(t *testing.T) {
	c := NewCookie("foo", "bar")
	if got, want := c.Name(), "foo"; got != want {
		t.Errorf("c.Name() got: %v want: %v", got, want)
	}
}

func TestCookieValue(t *testing.T) {
	c := NewCookie("foo", "bar")
	if got, want := c.Value(), "bar"; got != want {
		t.Errorf("c.Value() got: %v want: %v", got, want)
	}
}
