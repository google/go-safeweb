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

// Config represents a configuration passed to the linter.
type Config struct {
	Imports   []BannedApi `json:"imports"`
	Functions []BannedApi `json:"functions"`
}

type BannedApi struct {
	Name       string      `json:"name"` // fully qualified identifier name
	Msg        string      `json:"msg"`  // additional information e.g. rationale for banning
	Exemptions []Exemption `json:"exemptions"`
}

type Exemption struct {
	Justification string `json:"justification"`
	AllowedDir    string `json:"allowedDir"`
}

// ReadConfigs reads banned Apis from all files.
func ReadConfigs(files []string) (*Config, error) {
	var imports []BannedApi
	var fns []BannedApi

	for _, file := range files {
		config, err := readCfg(file)
		if err != nil {
			return nil, err
		}

		imports = append(imports, config.Imports...)
		fns = append(fns, config.Functions...)
	}

	return &Config{Imports: imports, Functions: fns}, nil
}

func readCfg(filename string) (*Config, error) {
	f, err := openFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return decodeCfg(f)
}

func openFile(filename string) (*os.File, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, errors.New("file does not exist")
	}
	if info.IsDir() {
		return nil, errors.New("file is a directory")
	}

	return os.Open(filename)
}

func decodeCfg(f *os.File) (*Config, error) {
	var cfg Config
	err := json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
