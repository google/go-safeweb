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
	parsed bool
}

// MultipartForm TODO
type MultipartForm struct {
	form Form
	file map[string][]*multipart.FileHeader
}

// Int TODO
func (f *Form) Int(paramName string, defaultValue int) int {
	const fnInt = "Int"
	if f.err != nil {
		return defaultValue
	}
	if !f.parsed {
		panic(fnInt + ": form has not been parsed")
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New(fnInt + ": no value found for key " + paramName)
		return defaultValue
	}
	paramVal, err := strconv.Atoi(vals[0])
	if err != nil {
		f.err = errors.New(fnInt + ": " + err.Error())
		return defaultValue
	}
	return paramVal
}

// Uint TODO
func (f *Form) Uint(paramName string, defaultValue uint64) uint64 {
	const fnUint = "Uint"
	if f.err != nil {
		return defaultValue
	}
	if !f.parsed {
		panic(fnUint + ": form has not been parsed")
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New(fnUint + ": no value found for key " + paramName)
		return defaultValue
	}
	paramVal, err := strconv.ParseUint(vals[0], 10, 0)
	if err != nil {
		f.err = errors.New(fnUint + ": " + err.Error())
		return defaultValue
	}
	return paramVal
}
func (f *Form) String(paramName string, defaultValue string) string {
	const fnString = "String"
	if f.err != nil {
		return defaultValue
	}
	if !f.parsed {
		panic(fnString + ": form has not been parsed")
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New(fnString + ": no value found for key " + paramName)
		return defaultValue
	}
	return vals[0]
}

// Float64 TODO
func (f *Form) Float64(paramName string, defaultValue float64) float64 {
	const fnFloat64 = "Float64"
	if f.err != nil {
		return defaultValue
	}
	if !f.parsed {
		panic(fnFloat64 + ": form has not been parsed")
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New(fnFloat64 + "no value found for key " + paramName)
		return defaultValue
	}
	paramVal, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		f.err = errors.New(fnFloat64 + ": " + err.Error())
		return defaultValue
	}
	return paramVal
}

// Bool TODO
func (f *Form) Bool(paramName string, defaultValue bool) bool {
	const fnBool = "Bool"
	if f.err != nil {
		return defaultValue
	}
	if !f.parsed {
		panic(fnBool + ": form has not been parsed")
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New(fnBool + ": no value found for key " + paramName)
		return defaultValue
	}
	if vals[0] != "true" {
		if vals[0] != "false" {
			f.err = errors.New(fnBool + ": values of form parameter " + paramName + " not a boolean")
		}
		return false
	}
	return true
}

// Slice TODO
func (f *Form) Slice(slice interface{}, paramName string) {
	const fnSlice = "Slice"
	if f.err != nil {
		slice = nil
		return
	}
	if !f.parsed {
		panic(fnSlice + ": form has not been parsed")
	}
	vals, ok := f.values[paramName]
	if !ok {
		f.err = errors.New(fnSlice + ": no value found for key " + paramName)
		slice = nil
	}
	len := len(vals)
	switch data := slice.(type) {
	case *[]string:
		res := *data
		res = make([]string, 0, len)
		for _, x := range vals {
			res = append(res, x)
		}
	case *[]int:
		res := *data
		res = make([]int, 0, len)
		for _, x := range vals {
			x, err := strconv.Atoi(x)
			if err != nil {
				f.err = errors.New(fnSlice + ": " + err.Error())
				slice = nil
				return
			}
			res = append(res, x)
		}
	case *[]uint64:
		res := *data
		res = make([]uint64, 0, len)
		for _, x := range vals {
			x, err := strconv.ParseUint(x, 10, 0)
			if err != nil {
				f.err = errors.New(fnSlice + ": " + err.Error())
				slice = nil
				return
			}
			res = append(res, x)
		}
	case *[]float64:
		res := *data
		res = make([]float64, 0, len)
		for _, x := range vals {
			x, err := strconv.ParseFloat(x, 64)
			if err != nil {
				f.err = errors.New(fnSlice + ": " + err.Error())
				slice = nil
				return
			}
			res = append(res, x)
		}
	case *[]bool:
		res := *data
		res = make([]bool, 0, len)
		for _, x := range vals {
			if x != "true" {
				if x != "false" {
					f.err = errors.New(fnSlice + ": " + ": values of form parameter " + paramName + " not a boolean")
					slice = nil
					return
				}
				res = append(res, false)
				continue
			}
			res = append(res, true)
		}
	default:
		f.err = errors.New(fnSlice + ": slice type not supported")
		slice = nil
	}

	return
}

// Error TODO
func (f *Form) Error() error {
	return f.err
}
