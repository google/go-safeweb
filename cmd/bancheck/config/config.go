package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// Config struct which contains an array of banned imports and function calls
type Config struct {
	Imports   []BannedImport   `json:"imports"`
	Functions []BannedFunction `json:"functions"`
}

// BannedImport struct which contains a fully qualified import name,
// message with additional information and a list of exemptions.
type BannedImport struct {
	Name       string      `json:"name"`
	Msg        string      `json:"msg"`
	Exemptions []Exemption `json:"exemptions"`
}

// BannedFunction struct which contains a fully qualified function name,
// message with additional information and a list of exemptions.
type BannedFunction struct {
	Name       string      `json:"name"`
	Msg        string      `json:"msg"`
	Exemptions []Exemption `json:"exemptions"`
}

// Exemption struct which contains a justification and a path to allowed directory.
type Exemption struct {
	Justification string `json:"justification"`
	AllowedDir    string `json:"allowedDir"`
}

// Read reads configs from all files and concatenates them into one object.
func Read(files []string) (*Config, error) {
	imports := make([]BannedImport, 0)
	functions := make([]BannedFunction, 0)

	for _, file := range files {
		config, err := unmarshalCfg(file)
		if err != nil {
			return nil, err
		}

		imports = append(imports, config.Imports...)
		functions = append(functions, config.Functions...)
	}

	return &Config{Imports: imports, Functions: functions}, nil
}

// unmarshalCfg reads JSON object from a file and converts it to a Config struct.
func unmarshalCfg(file string) (*Config, error) {
	if !fileExists(file) {
		return nil, errors.New("file does not exist or is a directory")
	}

	cfg, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer cfg.Close()

	var config Config
	bytes, _ := ioutil.ReadAll(cfg)
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
