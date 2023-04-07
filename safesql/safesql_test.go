// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safesql

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestTrustedSQLConcat(t *testing.T) {
	var tests = []struct {
		name string
		ss   []TrustedSQLString
		want TrustedSQLString
	}{
		{name: "nothing"},
		{
			name: "one word",
			ss:   []TrustedSQLString{New("foo")},
			want: New("foo"),
		},
		{
			name: "two words",
			ss:   []TrustedSQLString{New("foo"), New("ffa")},
			want: New("fooffa"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrustedSQLStringConcat(tt.ss...)
			if got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}
func TestTrustedSQLJoin(t *testing.T) {
	var tests = []struct {
		name string
		ss   []TrustedSQLString
		sep  TrustedSQLString
		want TrustedSQLString
	}{
		{name: "nothing"},
		{
			name: "one word",
			ss:   []TrustedSQLString{New("foo")},
			sep:  New("bar"),
			want: New("foo"),
		},
		{
			name: "two words",
			ss:   []TrustedSQLString{New("foo"), New("ffa")},
			sep:  New("bar"),
			want: New("foobarffa"),
		},
		{
			name: "empty sep",
			ss:   []TrustedSQLString{New("foo"), New("ffa")},
			want: New("fooffa"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrustedSQLStringJoin(tt.ss, tt.sep)
			if got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestTrustedSQLStringSplit(t *testing.T) {
	var tests = []struct {
		name string
		in   TrustedSQLString
		sep  TrustedSQLString
		want []TrustedSQLString
	}{
		{name: "nothing"},
		{
			name: "no sep",
			in:   New("foo"),
			sep:  New("bar"),
			want: []TrustedSQLString{New("foo")},
		},
		{
			name: "two words",
			in:   New("foobarffa"),
			sep:  New("bar"),
			want: []TrustedSQLString{New("foo"), New("ffa")},
		},
		{
			name: "empty sep",
			in:   New("foo"),
			want: []TrustedSQLString{New("f"), New("o"), New("o")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrustedSQLStringSplit(tt.in, tt.sep)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty(), cmp.AllowUnexported(TrustedSQLString{})); diff != "" {
				t.Errorf("-want +got: %s", diff)
			}
		})
	}
}
