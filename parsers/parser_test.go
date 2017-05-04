package parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Case struct {
	Description  string
	Input        string
	Expected     ParserResult
	ExpectedRest []string
	RunForAll    bool
}

type CaseGroup struct {
	name  string
	funct func(lines []string) (ParserResult, []string)
	cases []Case
}

var ParserTests = []CaseGroup{
	CaseGroup{"assets", ParseAssets, assetListTestCases},
	CaseGroup{"cargo_scans", ParseCargoScan, cargoScanTestCases},
	CaseGroup{"contracts", ParseContract, contractTestCases},
	CaseGroup{"dscan", ParseDScan, dscanTestCases},
}

func TestParsers(rt *testing.T) {
	for _, group := range ParserTests {
		for _, c := range group.cases {
			rt.Run(group.name+":"+c.Description, func(t *testing.T) {
				result, rest := group.funct(strings.Split(c.Input, "\n"))
				assert.Equal(t, c.Expected, result, "results should be the same")
				assert.Equal(t, c.ExpectedRest, rest, "the rest should be the same")
			})
		}
	}

	for _, group := range ParserTests {
		for _, c := range group.cases {
			if !c.RunForAll {
				continue
			}
			rt.Run("AllParser_"+group.name+":"+c.Description, func(t *testing.T) {
				result, rest := AllParser(strings.Split(c.Input, "\n"))
				assert.Equal(t, &MultiParserResult{results: []ParserResult{c.Expected}}, result, "results should be the same")
				assert.Equal(t, c.ExpectedRest, rest, "the rest should be the same")
			})
		}
	}
}
