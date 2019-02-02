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
	{"assets", ParseAssets, assetListTestCases},
	{"cargo_scans", ParseCargoScan, cargoScanTestCases},
	{"contracts", ParseContract, contractTestCases},
	{"dscan", ParseDScan, dscanTestCases},
	{"listing", ParseListing, listingTestCases},
	{"eft", ParseEFT, eftTestCases},
	{"fitting", ParseFitting, fittingTestCases},
	{"industry", ParseIndustry, industryTestCases},
	{"loot_history", ParseLootHistory, lootHistoryTestCases},
	{"pi", ParsePI, piTestCases},
	{"survey_scanner", ParseSurveyScan, surveyScannerTestCases},
	{"view_contents", ParseViewContents, viewContentsTestCases},
	{"wallet", ParseWallet, walletTestCases},
	{"killmail", ParseKillmail, killmailTestCases},
	{"mining_ledger", ParseMiningLedger, miningLedgerTestCases},
	{"moon_ledger", ParseMoonLedger, moonLedgerTestCases},
	{"compare", ParseCompare, compareTestCases},
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

				expectedResult := &MultiParserResult{Results: []ParserResult{c.Expected}}
				if c.Expected == nil {
					expectedResult = &MultiParserResult{Results: nil}
				}
				assert.Equal(t, expectedResult, result, "results should be the same")
				assert.Equal(t, c.ExpectedRest, rest, "the rest should be the same")
			})
		}
	}
}
