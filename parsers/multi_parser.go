package parsers

import "sort"

var AllParser = NewMultiParser(AllParsers)

type MultiParserResult struct {
	Results []ParserResult
}

func (r *MultiParserResult) Name() string {
	return "multi"
}

func (r *MultiParserResult) Lines() []int {
	lines := make([]int, 0)
	for _, r := range r.Results {
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
				if len(left) == 0 {
					break
				}
				var result ParserResult
				result, left = parser(left)
				if result != nil && len(result.Lines()) > 0 {
					multiParserResult.Results = append(multiParserResult.Results, result)
				}
			}
			return multiParserResult, left
		})
}
