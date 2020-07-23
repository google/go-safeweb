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
	"mime/multipart"
	"strconv"
)

// Form contains parsed data either from URL's query or from
// form parameters, part of the body of POST, PATCH or PUT requests that are not
// multipart requests. The fields are only available after parsing the form,
// through getter functions that specify the type. If parsing failed, Form will
// be set to nil. Field err will only be set if
// an error occurs when the user tries to access a parameter.
type Form struct {
	values map[string][]string
	err    error
}

// Int checks whether key paramKey maps to any query or form parameter
// values. In case it does, it will try to convert the first value to an integer
// and return it. If there are no values associated with paramKey, it will
// return the default value. If the first value is not an integer, it will
// return the default value and set the Form error field.
func (f *Form) Int(paramKey string, defaultValue int) int {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramKey]
	if !ok {
		return defaultValue
	}
	paramVal, err := strconv.Atoi(vals[0])
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// Uint checks whether key paramKey maps to any query or form parameter
// values. In case it does, it will try to convert the first value to an
// unsigned integer and return it. If there are no values associated with
// paramKey, it will return the default value. If the first value is not an
// unsigned integer, it will return the default value and set the Form
// error field.
func (f *Form) Uint(paramKey string, defaultValue uint64) uint64 {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramKey]
	if !ok {
		return defaultValue
	}
	paramVal, err := strconv.ParseUint(vals[0], 10, 0)
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// String checks whether key paramKey maps to any query or form parameter
// values. In case it does, it will return the first value. If it doesn't, it
// will return the default value.
func (f *Form) String(paramKey string, defaultValue string) string {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramKey]
	if !ok {
		return defaultValue
	}
	return vals[0]
}

// Float64 checks whether key paramKey maps to any query or form parameter
// values. In case it does, it will try to convert the first value to a float
// and return it. If there are no values associated with paramKey, it will
// return the default value. If the first value is not a float, it will return
// the default value and set the Form error field.
func (f *Form) Float64(paramKey string, defaultValue float64) float64 {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramKey]
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

// Bool checks whether key paramKey maps to any query  or form parameter
// values. In case it does, it will try to convert the first value to a boolean
// and return it. If there are no values associated with paramKey, it will
// return the default value. If the first value is not a boolean, it will return
// the default value and set the Form error field.
func (f *Form) Bool(paramKey string, defaultValue bool) bool {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramKey]
	if !ok {
		return defaultValue
	}
	switch vals[0] {
	case "true":
		return true
	case "false":
		return false
	default:
		f.err = fmt.Errorf("values of form parameter %q not a boolean", paramKey)
	}
	return false
}

func clearSlice(slicePtr interface{}) error {
	switch vs := slicePtr.(type) {
	case *[]string:
		*vs = nil
	case *[]int:
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

// Slice checks whether key paramKey maps to any query or form parameters. If it
// does, it will try to convert them to the type of slice elements slicePtr
// points to. If there are no values associated with paramKey, it will clear the
// slice. If type conversion fails at any point, the Form error field will be
// set and the slice will be cleared.
func (f *Form) Slice(slicePtr interface{}, paramKey string) {
	if f.err != nil {
		f.err = clearSlice(slicePtr)
		return
	}
	mapVals, ok := f.values[paramKey]
	if !ok {
		f.err = clearSlice(slicePtr)
		return
	}
	switch values := slicePtr.(type) {
	case *[]string:
		res := make([]string, 0, len(mapVals))
		for _, x := range mapVals {
			res = append(res, x)
		}
		*values = res
	case *[]int:
		res := make([]int, 0, len(mapVals))
		for _, x := range mapVals {
			x, err := strconv.Atoi(x)
			if err != nil {
				f.err = err
				*values = nil
				return
			}
			res = append(res, x)
			*values = res
		}
	case *[]uint64:
		res := make([]uint64, 0, len(mapVals))
		for _, x := range mapVals {
			x, err := strconv.ParseUint(x, 10, 0)
			if err != nil {
				f.err = err
				*values = nil
				slicePtr = values
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
				f.err = fmt.Errorf("values of form parameter %q not a boolean", paramKey)
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

// Err returns the value of the Form error field. This will be nil unless an
// error occured while accessing a parsed form value.
func (f *Form) Err() error {
	return f.err
}

// MultipartForm extends the Form structure to define a POST, PATCH or PUT
// request that has Content-Type: multipart/form-data. Its fields are only
// available after parsing the form, through getter functions that specify the
// type.
type MultipartForm struct {
	Form
	file map[string][]*multipart.FileHeader
}

// TODO(@mihalimara22): Create getters and tests for the `file` field in MultipartForm
