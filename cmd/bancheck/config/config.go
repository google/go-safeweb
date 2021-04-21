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
package config

import (
	"encoding/json"
	"errors"
	"os"
)

type BannedApi struct {
	Name       string      `json:"name"` // fully qualified identifier name
	Msg        string      `json:"msg"`  // additional information e.g. rationale for banning
	Exemptions []Exemption `json:"exemptions"`
}

// BannedApis is a map of identifier names to a list of all corresponding
// BannedApi entries with additional information.
type BannedApis map[string][]BannedApi

type Exemption struct {
	Justification string `json:"justification"`
	AllowedDir    string `json:"allowedDir"`
}

// ReadBannedImports reads banned imports from all config files
// and concatenates them into one object.
func ReadBannedImports(files []string) (BannedApis, error) {
	imports := make(BannedApis)

	for _, file := range files {
		config, err := unmarshalCfg(file)
		if err != nil {
			return nil, err
		}

		for _, i := range config.Imports {
			imports[i.Name] = append(imports[i.Name], i)
		}
	}

	return imports, nil
}

// ReadBannedFunctions reads banned function calls from all config files
// and concatenates them into a map.
func ReadBannedFunctions(files []string) (BannedApis, error) {
	fns := make(BannedApis)

	for _, file := range files {
		config, err := unmarshalCfg(file)
		if err != nil {
			return nil, err
		}

		for _, fn := range config.Functions {
			fns[fn.Name] = append(fns[fn.Name], fn)
		}
	}

	return fns, nil
}

// config represents contents of a configuration file passed to the linter.
type config struct {
	Imports   []BannedApi `json:"imports"`
	Functions []BannedApi `json:"functions"`
}

// unmarshalCfg reads JSON object from a file and converts it to a bannedAPIs struct.
func unmarshalCfg(filename string) (*config, error) {
	f, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg config
	d := json.NewDecoder(f)
	err = d.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func readFile(filename string) (*os.File, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, errors.New("file does not exist")
	}
	if info.IsDir() {
		return nil, errors.New("file is a directory")
	}

	return os.Open(filename)
}
