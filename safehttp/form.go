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
// the body of POST, PATCH or PUT requests that are not multipart requests.
type Form struct {
	values map[string][]string
	err    error
}

// Int64 returns the first form parameter value. If the first value is not a
// valid int64, the defaultValue is returned instead and an error is set
// (retrievable by Err()).
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

// Uint64 returns the first form parameter value. If the first value is not a
// valid uint64, the defaultValue is returned instead and an error is set
// (retrievable by Err()).
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

// String returns the first form parameter value. If the first value is not a
// valid string, the defaultValue is returned instead and an error is set
// (retrievable by Err()).
func (f *Form) String(param string, defaultValue string) string {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	return vals[0]
}

// Float64 returns the first form parameter value. If the first value is not a
// valid float64, the defaultValue is returned instead and an error is set
// (retrievable by Err()).
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

// Bool returns the first form parameter value. If the first value is not a
// valid bool, the defaultValue is returned instead and an error is set
// (retrievable by Err()).
func (f *Form) Bool(param string, defaultValue bool) bool {
	vals, ok := f.values[param]
	if !ok {
		return defaultValue
	}
	b, err := strconv.ParseBool(vals[0])
	if err != nil {
		f.err = err
		return defaultValue
	}
	return b
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

// Slice returns the form parameter values. If the values don't have the same
// type, slicePtr will point to a nil slice instead and an error is set
// (retrievable by Err()). This function should be used in case a form parameter
// maps to multiple values.
//
// TODO(mihalimara22): Simplify this function to avoid duplicate logic
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
			b, err := strconv.ParseBool(x)
			if err != nil {
				f.err = err
				*values = nil
				return
			}
			res = append(res, b)
		}
		*values = res
	default:
		f.err = clearSlice(slicePtr)
	}
}

// Err returns nil unless an error occurred while accessing a parsed form value.
// Calling this method will return the last error that occurred while parsing
// form values.
func (f *Form) Err() error {
	return f.err
}

// MultipartForm extends the Form structure to define a POST, PATCH or PUT
// request that has Content-Type: multipart/form-data.
type MultipartForm struct {
	Form
	// TODO(mihalimara22): add support for form files in the MultipartForm
}
