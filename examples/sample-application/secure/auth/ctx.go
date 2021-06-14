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

// +build go1.16

package auth

import (
	"context"

	"github.com/google/go-safeweb/safehttp"
)

type ctxKey string

const (
	userCtx       ctxKey = "user"
	changeSessCtx ctxKey = "change"
)

type sessionAction string

const (
	clearSess sessionAction = "clear"
	setSess   sessionAction = "set"
)

func putSessionAction(ctx context.Context, action sessionAction) {
	m := safehttp.FlightValues(ctx)
	m.Put(changeSessCtx, action)
}

func ctxSessionAction(ctx context.Context) sessionAction {
	m := safehttp.FlightValues(ctx)
	v := m.Get(changeSessCtx)
	action, ok := v.(sessionAction)
	if !ok {
		return ""
	}
	return action
}

func putUser(ctx context.Context, user string) {
	m := safehttp.FlightValues(ctx)
	m.Put(userCtx, user)
}

func ctxUser(ctx context.Context) string {
	m := safehttp.FlightValues(ctx)
	v := m.Get(userCtx)
	user, ok := v.(string)
	if !ok {
		return ""
	}
	return user
}
