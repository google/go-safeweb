package analyzers

import (
	"errors"
	"flag"
	"strings"

	"github.com/google/go-safeweb/cmd/bancheck/config"
	"golang.org/x/tools/go/analysis"
)

func BannedImports(pass *analysis.Pass) (map[string]config.BannedImport, error) {
	cfg, err := readConfig(pass)
	if err != nil {
		return nil, err
	}

	return cfg.BannedImports(), nil
}

func BannedFunctions(pass *analysis.Pass) (map[string]config.BannedFunction, error) {
	cfg, err := readConfig(pass)
	if err != nil {
		return nil, err
	}

	return cfg.BannedFunctions(), nil
}

func readConfig(pass *analysis.Pass) (*config.Config, error) {
	cfgFiles := pass.Analyzer.Flags.Lookup("config").Value.(flag.Getter).Get().(string)
	if cfgFiles == "" {
		return nil, errors.New("please, provide a config file")
	}

	return config.Read(strings.Split(cfgFiles, ","))
}
