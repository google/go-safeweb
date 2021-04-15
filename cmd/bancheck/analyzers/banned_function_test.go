package analyzers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBannedFunctionAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), BannedFunctionAnalyzer)
}
