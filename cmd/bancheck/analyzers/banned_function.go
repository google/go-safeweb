package analyzers

import (
	"errors"

	"golang.org/x/tools/go/analysis"
)

var BannedFunctionAnalyzer = &analysis.Analyzer{
	Name: "bannedfunction",
	Doc:  "Checks for usage of banned functions and reports them",
	Run:  checkBannedFunctions,
}

func checkBannedFunctions(pass *analysis.Pass) (interface{}, error) {
	return nil, errors.New("not implemented yet")
}
