// Copyright 2022 Google LLC
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

package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/google/go-safeweb/safehttp"
)

func TestPutSessionAction(t *testing.T) {
	req := safehttp.NewIncomingRequest(httptest.NewRequest(safehttp.MethodGet, "/spaghetti", nil))
	ctx := req.Context()

	putSessionAction(ctx, setSess)

	if v := safehttp.FlightValues(ctx).Get(changeSessKey); v != setSess {
		t.Errorf("putSessionAction(), got: %q, want: %q ", v, setSess)
	}
}

func TestCtxSessionAction(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  sessionAction
	}{
		{
			name:  "basic session action set/get",
			value: setSess,
			want:  setSess,
		},
		{
			name:  "no sessionAction value, return empty string",
			value: "unexpected",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := safehttp.NewIncomingRequest(httptest.NewRequest(safehttp.MethodGet, "/", nil))
			ctx := req.Context()

			safehttp.FlightValues(ctx).Put(changeSessKey, tt.value)

			if got := ctxSessionAction(ctx); got != tt.want {
				t.Errorf("ctxSessionAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPutUser(t *testing.T) {
	req := safehttp.NewIncomingRequest(httptest.NewRequest(safehttp.MethodGet, "/spaghetti", nil))
	ctx := req.Context()
	want := "someoneName"

	putUser(ctx, want)

	if v := safehttp.FlightValues(ctx).Get(userKey); v != want {
		t.Errorf("putUser(), got: %q, want: %q ", v, setSess)
	}
}

func TestCtxUser(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{
			name:  "basic session user set/get",
			value: "safeUser",
			want:  "safeUser",
		},
		{
			name:  "no string value, return empty string",
			value: 42,
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := safehttp.NewIncomingRequest(httptest.NewRequest(safehttp.MethodGet, "/", nil))
			ctx := req.Context()

			safehttp.FlightValues(ctx).Put(userKey, tt.value)

			if got := ctxUser(ctx); got != tt.want {
				t.Errorf("ctxUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
