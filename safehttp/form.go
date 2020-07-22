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
	"errors"
	"mime/multipart"
	"strconv"
)

// Form TODO
type Form struct {
	values map[string][]string
	err    error
}

// MultipartForm extends the Form structure to define a POST, PATCH or PUT
// request that is multipart. Its fields are only available after parsing the
// form, through getter functions that specify the type.
type MultipartForm struct {
	*Form
	file map[string][]*multipart.FileHeader
}

// Int TODO
func (f *Form) Int(paramName string, defaultValue int) int {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New("no value found for key " + paramName)
		return defaultValue
	}
	paramVal, err := strconv.Atoi(vals[0])
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// Uint TODO
func (f *Form) Uint(paramName string, defaultValue uint64) uint64 {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New("no value found for key " + paramName)
		return defaultValue
	}
	paramVal, err := strconv.ParseUint(vals[0], 10, 0)
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}
func (f *Form) String(paramName string, defaultValue string) string {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New("no value found for key " + paramName)
		return defaultValue
	}
	return vals[0]
}

// Float64 TODO
func (f *Form) Float64(paramName string, defaultValue float64) float64 {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New("no value found for key " + paramName)
		return defaultValue
	}
	paramVal, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		f.err = err
		return defaultValue
	}
	return paramVal
}

// Bool TODO
func (f *Form) Bool(paramName string, defaultValue bool) bool {
	if f.err != nil {
		return defaultValue
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New("no value found for key " + paramName)
		return defaultValue
	}
	if vals[0] != "true" {
		if vals[0] != "false" {
			f.err = errors.New("values of form parameter " + paramName + " not a boolean")
		}
		return false
	}
	return true
}

// Slice TODO
func (f *Form) Slice(slice interface{}, paramName string) {
	mapVals, ok := f.values[paramName]
	switch values := slice.(type) {
	case *[]string:
		if f.err != nil {
			*values = nil
			return
		}
		if !ok {
			f.err = errors.New("no value found for key " + paramName)
			*values = nil
		}
		res := make([]string, 0, len(mapVals))
		for _, x := range mapVals {
			res = append(res, x)
		}
		*values = res
	case *[]int:
		if f.err != nil {
			*values = nil
			return
		}
		if !ok {
			f.err = errors.New("no value found for key " + paramName)
			*values = nil
		}
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
		if f.err != nil {
			*values = nil
			return
		}
		if !ok {
			f.err = errors.New("no value found for key " + paramName)
			*values = nil
		}
		res := make([]uint64, 0, len(mapVals))
		for _, x := range mapVals {
			x, err := strconv.ParseUint(x, 10, 0)
			if err != nil {
				f.err = err
				*values = nil
				return
			}
			res = append(res, x)
		}
		*values = res
	case *[]float64:
		if f.err != nil {
			*values = nil
			return
		}
		if !ok {
			f.err = errors.New("no value found for key " + paramName)
			*values = nil
		}
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
		if f.err != nil {
			*values = nil
			return
		}
		if !ok {
			f.err = errors.New("no value found for key " + paramName)
			*values = nil
		}
		res := make([]bool, 0, len(mapVals))
		for _, x := range mapVals {
			if x != "true" {
				if x != "false" {
					f.err = errors.New(": values of form parameter " + paramName + " not a boolean")
					*values = nil
					return
				}
				res = append(res, false)
				continue
			}
			res = append(res, true)
		}
		*values = res
	default:
		f.err = errors.New("slice type not supported")
	}
	return
}

// Error TODO
func (f *Form) Error() error {
	return f.err
}
