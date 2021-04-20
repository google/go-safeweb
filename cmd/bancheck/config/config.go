package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// Config struct which contains an array of banned imports and function calls
type config struct {
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

// BannedImports is a map of import names to a list of BannedImport
// entries that define additional information.
type BannedImports map[string][]BannedImport

// BannedFunction struct which contains a fully qualified function name,
// message with additional information and a list of exemptions.
type BannedFunction struct {
	Name       string      `json:"name"`
	Msg        string      `json:"msg"`
	Exemptions []Exemption `json:"exemptions"`
}

// BannedFunctions is a map of fully qualified function names
// to a list of BannedFunction entries that define additional information.
type BannedFunctions map[string][]BannedFunction

// Exemption struct which contains a justification and a path to allowed directory.
type Exemption struct {
	Justification string `json:"justification"`
	AllowedDir    string `json:"allowedDir"`
}

// ReadBannedImports reads banned imports from all config files
// and concatenates them into one object.
func ReadBannedImports(files []string) (BannedImports, error) {
	imports := make(BannedImports)

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
func ReadBannedFunctions(files []string) (BannedFunctions, error) {
	fns := make(BannedFunctions)

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

// unmarshalCfg reads JSON object from a file and converts it to a Config struct.
func unmarshalCfg(file string) (*config, error) {
	if !fileExists(file) {
		return nil, errors.New("file does not exist or is a directory")
	}

	cfg, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer cfg.Close()

	var config config
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
