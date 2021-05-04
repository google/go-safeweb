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

package bannedapi

import (
	"errors"
	"flag"
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/google/go-safeweb/cmd/bancheck/config"
	"golang.org/x/tools/go/analysis"
)

// NewAnalyzer returns an analyzer that checks for usage of banned APIs.
func NewAnalyzer() *analysis.Analyzer {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.String("configs", "", "Config files with banned APIs separated by a comma")

	a := &analysis.Analyzer{
		Name:  "bannedAPI",
		Doc:   "Checks for usage of banned APIs",
		Run:   checkBannedAPIs,
		Flags: *fs,
	}

	return a
}

func checkBannedAPIs(pass *analysis.Pass) (interface{}, error) {
	cfgFiles := pass.Analyzer.Flags.Lookup("configs").Value.String()
	if cfgFiles == "" {
		return nil, errors.New("missing config files")
	}

	cfg, err := config.ReadConfigs(strings.Split(cfgFiles, ","))
	if err != nil {
		return nil, err
	}

	checkBannedImports(pass, bannedAPIMap(cfg.Imports))
	checkBannedFunctions(pass, bannedAPIMap(cfg.Functions))

	return nil, nil
}

func checkBannedImports(pass *analysis.Pass, bannedImports map[string][]config.BannedAPI) (interface{}, error) {
	for _, f := range pass.Files {
		for _, i := range f.Imports {
			importName := strings.Trim(i.Path.Value, "\"")
			err := reportIfBanned(importName, bannedImports, i.Pos(), pass)
			if err != nil {
				return false, err
			}
		}
	}
	return nil, nil
}

func checkBannedFunctions(pass *analysis.Pass, bannedFns map[string][]config.BannedAPI) (interface{}, error) {
	for id, obj := range pass.TypesInfo.Uses {
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}

		fnName := fmt.Sprintf("%s.%s", fn.Pkg().Path(), fn.Name())
		err := reportIfBanned(fnName, bannedFns, id.Pos(), pass)
		if err != nil {
			return false, err
		}
	}
	return nil, nil
}

func reportIfBanned(apiName string, bannedAPIs map[string][]config.BannedAPI, position token.Pos, pass *analysis.Pass) error {
	bannedAPICfgs, isBanned := bannedAPIs[apiName]
	if !isBanned {
		return nil
	}
	pkgAllowed, err := isPkgAllowed(pass.Pkg, bannedAPICfgs)
	if err != nil {
		return err
	}
	if pkgAllowed {
		return nil
	}
	for _, bannedAPICfg := range bannedAPICfgs {
		pass.Report(analysis.Diagnostic{
			Pos:     position,
			Message: fmt.Sprintf("Banned API found %q. Additional info: %s", apiName, bannedAPICfg.Msg),
		})
	}
	return nil
}

// isPkgAllowed checks if the Go package should be exempted from reporting banned API usages.
func isPkgAllowed(pkg *types.Package, bannedAPI []config.BannedAPI) (bool, error) {
	for _, fn := range bannedAPI {
		for _, e := range fn.Exemptions {
			match, err := filepath.Match(e.AllowedPkg, pkg.Path())
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

// bannedAPIMap builds a mapping of fully qualified API name to a list of
// all its config.BannedAPI entries.
func bannedAPIMap(bannedAPIs []config.BannedAPI) map[string][]config.BannedAPI {
	m := make(map[string][]config.BannedAPI)
	for _, API := range bannedAPIs {
		m[API.Name] = append(m[API.Name], API)
	}
	return m
}
