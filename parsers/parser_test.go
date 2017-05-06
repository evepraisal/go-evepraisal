package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Case struct {
	Description  string
	Input        string
	Expected     ParserResult
	ExpectedRest Input
	RunForAll    bool
}

type CaseGroup struct {
	name  string
	funct func(input Input) (ParserResult, Input)
	cases []Case
}

var ParserTests = []CaseGroup{
	CaseGroup{"assets", ParseAssets, assetListTestCases},
	CaseGroup{"cargo_scans", ParseCargoScan, cargoScanTestCases},
	CaseGroup{"contracts", ParseContract, contractTestCases},
	CaseGroup{"dscan", ParseDScan, dscanTestCases},
	CaseGroup{"listing", ParseListing, listingTestCases},
	CaseGroup{"eft", ParseEFT, eftTestCases},
	CaseGroup{"fitting", ParseFitting, fittingTestCases},
	CaseGroup{"industry", ParseIndustry, industryTestCases},
	CaseGroup{"loot_history", ParseLootHistory, lootHistoryTestCases},
}

func TestParsers(rt *testing.T) {
	for _, group := range ParserTests {
		for _, c := range group.cases {
			rt.Run(group.name+":"+c.Description, func(t *testing.T) {
				result, rest := group.funct(StringToInput(c.Input))
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
				result, rest := AllParser(StringToInput(c.Input))
				assert.Equal(t, &MultiParserResult{results: []ParserResult{c.Expected}}, result, "results should be the same")
				assert.Equal(t, c.ExpectedRest, rest, "the rest should be the same")
			})
		}
	}
}
