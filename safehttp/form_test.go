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
	"math"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormValidInt64(t *testing.T) {
	tests := []struct {
		val  string
		want int64
	}{
		{val: "123", want: 123},
		{val: "9223372036854775807", want: math.MaxInt64},
		{val: "-1", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got := f.Int64("a", 0); got != tt.want {
				t.Errorf(`f.Int64("a", 0) got: %v want: %v`, got, tt.want)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestFormInvalidInt64(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{name: "Overflow", val: "9223372036854775810"},
		{name: "Not a number", val: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got, want := f.Int64("a", 0), int64(0); got != want {
				t.Errorf(`f.Int64("a", 0) got: %v want: %v`, got, want)
			}

			if err := f.Err(); err == nil {
				t.Error("f.Err() got: nil want: error")
			}
		})
	}
}

func TestFormValidUint64(t *testing.T) {
	tests := []struct {
		val  string
		want uint64
	}{
		{val: "123", want: 123},
		{val: "18446744073709551615", want: math.MaxUint64},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got := f.Uint64("a", 0); got != tt.want {
				t.Errorf(`f.Uint64("a", 0) got: %v want: %v`, got, tt.want)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestFormInvalidUint(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{name: "Negative", val: "-1"},
		{name: "Overflow", val: "18446744073709551630"},
		{name: "Not a number", val: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got, want := f.Uint64("a", 0), uint64(0); got != want {
				t.Errorf(`f.Uint64("a", 0) got: %v want: %v`, got, want)
			}

			if err := f.Err(); err == nil {
				t.Error("f.Err() got: nil want: error")
			}
		})
	}
}

func TestFormValidString(t *testing.T) {
	tests := []string{
		"b",
		"diavola",
		"ăȚâȘî",
		"\x64\x69\x61\x76\x6f\x6c\x61",
	}

	for _, val := range tests {
		t.Run(val, func(t *testing.T) {
			values := map[string][]string{"a": {val}}
			f := Form{values: values}

			if got := f.String("a", ""); got != val {
				t.Errorf(`f.String("a", 0) got: %v want: %v`, got, val)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestFormValidFloat64(t *testing.T) {
	tests := []struct {
		val  string
		want float64
	}{
		{
			val:  "1.234",
			want: 1.234,
		},
		{
			val:  strconv.FormatFloat(math.MaxFloat64, 'f', 6, 64),
			want: math.MaxFloat64,
		},
		{
			val:  strconv.FormatFloat(-math.SmallestNonzeroFloat64, 'f', 324, 64),
			want: -math.SmallestNonzeroFloat64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got := f.Float64("a", 0); got != tt.want {
				t.Errorf(`f.Float64("a", 0) got: %v want: %v`, got, tt.want)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestFormInvalidFloat64(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{name: "Not a float", val: "abc"},
		{name: "Overflow", val: "1.797693134862315708145274237317043567981e309"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got, want := f.Float64("a", 0), float64(0); got != want {
				t.Errorf(`f.Float64("a", 0) got: %v want: %v`, got, want)
			}

			if err := f.Err(); err == nil {
				t.Error("f.Err() got: nil want: error")
			}
		})
	}
}

func TestFormValidBool(t *testing.T) {
	tests := []struct {
		val  string
		want bool
	}{
		{val: "true", want: true},
		{val: "True", want: true},
		{val: "TRUE", want: true},
		{val: "t", want: true},
		{val: "T", want: true},
		{val: "1", want: true},
		{val: "false", want: false},
		{val: "False", want: false},
		{val: "FALSE", want: false},
		{val: "f", want: false},
		{val: "F", want: false},
		{val: "0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got := f.Bool("a", false); got != tt.want {
				t.Errorf(`f.Bool("a", 0) got: %v want: %v`, got, tt.want)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestFormInvalidBool(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{name: "Invalid casing", val: "TRuE"},
		{name: "Not a bool", val: "potato"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string][]string{"a": {tt.val}}
			f := Form{values: values}

			if got, want := f.Bool("a", false), false; got != want {
				t.Errorf(`f.Bool("a", 0) got: %v want: %v`, got, want)
			}

			if err := f.Err(); err == nil {
				t.Error("f.Err() got: nil want: error")
			}
		})
	}
}

func TestFormValidSlice(t *testing.T) {
	tests := []struct {
		name        string
		values      []string
		placeholder interface{}
		want        interface{}
	}{
		{
			name:        "Int64",
			values:      []string{"-8", "9", "-100"},
			placeholder: &[]int64{},
			want:        &[]int64{-8, 9, -100},
		},
		{
			name:        "Uint64",
			values:      []string{"8", "9", "10"},
			placeholder: &[]uint64{},
			want:        &[]uint64{8, 9, 10},
		},
		{
			name:        "String",
			values:      []string{"margeritta", "diavola", "calzone"},
			placeholder: &[]string{},
			want:        &[]string{"margeritta", "diavola", "calzone"},
		},
		{
			name:        "Float64",
			values:      []string{"1.3", "8.9", "-4.1"},
			placeholder: &[]float64{},
			want:        &[]float64{1.3, 8.9, -4.1},
		},
		{
			name:        "Bool",
			values:      []string{"t", "0", "TRUE"},
			placeholder: &[]bool{},
			want:        &[]bool{true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Form{values: map[string][]string{"x": tt.values}}

			f.Slice("x", tt.placeholder)
			if diff := cmp.Diff(tt.want, tt.placeholder); diff != "" {
				t.Errorf("f.Slice: diff (-want +got): \n%v", diff)
			}

			if err := f.Err(); err != nil {
				t.Errorf("f.Err() got: %v want: nil", err)
			}
		})
	}
}

func TestFormInvalidSlice(t *testing.T) {
	tests := []struct {
		name        string
		values      []string
		placeholder interface{}
	}{
		{
			name:        "Int64",
			values:      []string{"1", "abc", "1"},
			placeholder: &[]int64{},
		},
		{
			name:        "Uint64",
			values:      []string{"1", "abc", "-1"},
			placeholder: &[]uint64{},
		},
		{
			name:        "Float64",
			values:      []string{"1.3", "abc", "-4.1"},
			placeholder: &[]float64{},
		},
		{
			name:        "Bool",
			values:      []string{"t", "abc", "TRUE"},
			placeholder: &[]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Form{values: map[string][]string{"x": tt.values}}

			f.Slice("x", tt.placeholder)

			// TODO: add a check here that tt.placeholder is nil. I (grenfeldt@)
			// can't come up with a way of testing this in a table test.

			if err := f.Err(); err == nil {
				t.Error("f.Err() got: nil want: error")
			}
		})
	}
}

func TestFormSliceInvalidPointerType(t *testing.T) {
	f := Form{values: map[string][]string{"x": {"x"}}}
	f.Slice("x", &[]int16{})
	if err := f.Err(); err == nil {
		t.Error("f.Err() got: nil want: error")
	}
}

func TestFormUnknownParam(t *testing.T) {
	tests := []struct {
		name         string
		getFormValue func(f Form) interface{}
		want         interface{}
	}{
		{
			name: "Int64",
			getFormValue: func(f Form) interface{} {
				return f.Int64("x", -15)
			},
			want: int64(-15),
		},
		{
			name: "Uint64",
			getFormValue: func(f Form) interface{} {
				return f.Uint64("x", 15)
			},
			want: uint64(15),
		},
		{
			name: "String",
			getFormValue: func(f Form) interface{} {
				return f.String("x", "missing")
			},
			want: "missing",
		},
		{
			name: "Float64",
			getFormValue: func(f Form) interface{} {
				return f.Float64("x", 3.14)
			},
			want: 3.14,
		},
		{
			name: "Bool",
			getFormValue: func(f Form) interface{} {
				return f.Bool("x", true)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Form{}
			got := tt.getFormValue(f)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("tt.getFormValue(f) diff (-want +got): \n%v", diff)
			}
		})
	}
}

func TestFormSliceUnknownParam(t *testing.T) {
	f := Form{}
	p := []int64{1, 2}
	f.Slice("x", &p)
	if p != nil {
		t.Errorf(`f.Slice("x", &p) got p: %v want: nil`, p)
	}
	if err := f.Err(); err != nil {
		t.Errorf("f.Err() got: %v want: nil", err)
	}
}

func TestFormClearSliceString(t *testing.T) {
	x := []string{"xyz"}
	if err := clearSlice(&x); err != nil {
		t.Errorf("clearSlice(&x) got err: %v want: nil", err)
	}
	if x != nil {
		t.Errorf("clearSlice(&x) got x: %v want: nil", x)
	}
}

func TestFormClearSliceInt64(t *testing.T) {
	x := []int64{-1}
	if err := clearSlice(&x); err != nil {
		t.Errorf("clearSlice(&x) got err: %v want: nil", err)
	}
	if x != nil {
		t.Errorf("clearSlice(&x) got x: %v want: nil", x)
	}
}

func TestFormClearSliceUint64(t *testing.T) {
	x := []uint64{1}
	if err := clearSlice(&x); err != nil {
		t.Errorf("clearSlice(&x) got err: %v want: nil", err)
	}
	if x != nil {
		t.Errorf("clearSlice(&x) got x: %v want: nil", x)
	}
}

func TestFormClearSliceFloat64(t *testing.T) {
	x := []float64{3.14}
	if err := clearSlice(&x); err != nil {
		t.Errorf("clearSlice(&x) got err: %v want: nil", err)
	}
	if x != nil {
		t.Errorf("clearSlice(&x) got x: %v want: nil", x)
	}
}

func TestFormClearSliceBool(t *testing.T) {
	x := []bool{true}
	if err := clearSlice(&x); err != nil {
		t.Errorf("clearSlice(&x) got err: %v want: nil", err)
	}
	if x != nil {
		t.Errorf("clearSlice(&x) got x: %v want: nil", x)
	}
}

func TestFormClearSliceUnknownType(t *testing.T) {
	x := []int16{-1}
	if err := clearSlice(&x); err == nil {
		t.Error("clearSlice(&x) got: nil want: error")
	}
}
