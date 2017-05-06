package parsers

import "sort"

type ParserResult interface {
	Name() string
	Lines() []int
}

type Parser func(input Input) (ParserResult, Input)

var AllParser = NewMultiParser(
	[]Parser{
		ParseEFT,
		ParseFitting,
		ParseLootHistory,
		ParsePI,
		ParseViewContents,
		ParseWallet,
		ParseSurveyScan,
		ParseContract,
		ParseAssets,
		ParseIndustry,
		ParseCargoScan,
		ParseDScan,
		ParseListing,
	})

type MultiParserResult struct {
	results []ParserResult
}

func (r *MultiParserResult) Name() string {
	return "multi"
}

func (r *MultiParserResult) Lines() []int {
	lines := make([]int, 0)
	for _, r := range r.results {
		lines = append(lines, r.Lines()...)
	}
	sort.Ints(lines)
	return lines
}

func NewMultiParser(parsers []Parser) Parser {
	return Parser(
		func(input Input) (ParserResult, Input) {
			multiParserResult := &MultiParserResult{}
			left := input
			for _, parser := range parsers {
				var result ParserResult
				result, left = parser(left)
				if result != nil && len(result.Lines()) > 0 {
					multiParserResult.results = append(multiParserResult.results, result)
				}
				if len(left) == 0 {
					break
				}
			}
			return multiParserResult, left
		})
}
