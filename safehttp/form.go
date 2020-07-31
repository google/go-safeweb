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
	"fmt"
	"strconv"
)

// Form contains parsed data from form parameters, part of
// the body of POST, PATCH or PUT requests that are not multipart requests. The
// form values will only be available after parsing the form, and only through
// the getter functions.
type Form struct {
	values map[string][]string
	err    error
}

// Int64 checks whether key param maps to any form parameter
// values. In case it does, it will try to convert the first value to a valid
// int64 and return it. If there are no values associated with param, it
// will return the defaultValue value. If the first value is not an integer, it will
// return the defaultValue value and Err() will return the parsing error.
func (f *Form) Int64(param string, defaultValue int64) int64 {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	paramVal, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// Uint64 checks whether key param maps to any form parameter
// values. In case it does, it will try to convert the first valid
// uint64 and return it. If there are no values associated with
// param, it will return the defaultValue value. If the first value is not an
// unsigned integer, it will return the defaultValue value and set the Form
// error field.
func (f *Form) Uint64(param string, defaultValue uint64) uint64 {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	paramVal, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// String checks whether key param maps to any form parameter
// values. In case it does, it will return the first value. If it doesn't, it
// will return the defaultValue value.
func (f *Form) String(param string, defaultValue string) string {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	return vals[0]
}

// Float64 checks whether key param maps to any form parameter
// values. In case it does, it will try to convert the first value to a valid
// float64 and return it. If there are no values associated with param, it will
// return the defaultValue value. If the first value is not a float, it will return
// the defaultValue value and Err() will return the parsing error.
func (f *Form) Float64(param string, defaultValue float64) float64 {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	paramVal, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// Bool checks whether key param maps to any form parameter
// values. In case it does, it will try to convert the first value to a valid
// bool and return it. If there are no values associated with param, it will
// return the defaultValue value. If the first value is not a boolean, it will return
// the defaultValue value and Err() will return the parsing error.
func (f *Form) Bool(param string, defaultValue bool) bool {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	switch vals[0] {
	case "true":
		return true
	case "false":
		return false
	default:
		f.err = fmt.Errorf("values of form parameter %q not a boolean", param)
	}
	return false
}

func clearSlice(slicePtr interface{}) error {
	switch vs := slicePtr.(type) {
	case *[]string:
		*vs = nil
	case *[]int64:
		*vs = nil
	case *[]float64:
		*vs = nil
	case *[]uint64:
		*vs = nil
	case *[]bool:
		*vs = nil
	default:
		return fmt.Errorf("type not supported in Slice call: %T", vs)
	}
	return nil
}

// Slice checks whether key param maps to any form parameters. If it
// does, it will try to convert them to the type of slice elements slicePtr
// points to. If there are no values associated with param, it will clear the
// slice. If type conversion fails at any point, the Form error field will be
// set and the slice will be cleared.
func (f *Form) Slice(param string, slicePtr interface{}) {
	mapVals, ok := f.values[param]
	if !ok {
		f.err = clearSlice(slicePtr)
		return
	}
	switch values := slicePtr.(type) {
	case *[]string:
		res := make([]string, 0, len(mapVals))
		*values = append(res, mapVals...)
	case *[]int64:
		res := make([]int64, 0, len(mapVals))
		for _, x := range mapVals {
			x, err := strconv.ParseInt(x, 10, 64)
			if err != nil {
				f.err = err
				*values = nil
				return
			}
			res = append(res, x)
		}
		*values = res
	case *[]uint64:
		res := make([]uint64, 0, len(mapVals))
		for _, x := range mapVals {
			x, err := strconv.ParseUint(x, 10, 64)
			if err != nil {
				f.err = err
				*values = nil
				return
			}
			res = append(res, x)
		}
		*values = res
	case *[]float64:
		res := make([]float64, 0, len(mapVals))
		for _, x := range mapVals {
			x, err := strconv.ParseFloat(x, 64)
			if err != nil {
				f.err = err
				*values = nil
				return
			}
			res = append(res, x)
		}
		*values = res
	case *[]bool:
		res := make([]bool, 0, len(mapVals))
		for _, x := range mapVals {
			switch x {
			case "true":
				res = append(res, true)
			case "false":
				res = append(res, false)
			default:
				f.err = fmt.Errorf("values of form parameter %q not a boolean", param)
				*values = nil
				return
			}
		}
		*values = res

	default:
		f.err = clearSlice(slicePtr)
	}
	return
}

// Err returns nil unless an error occurred while accessing a parsed form value.
// Calling this method will return the last error that occurred while parsing
// form values.
func (f *Form) Err() error {
	return f.err
}

// MultipartForm extends the Form structure to define a POST, PATCH or PUT
// request that has Content-Type: multipart/form-data. Its fields are only
// available after parsing the form, through getter functions that specify the
// type.
type MultipartForm struct {
	Form
}
