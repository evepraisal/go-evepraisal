package parsers

import "sort"

// AllParser is a multi-parser that uses all of the default parsers
var AllParser = NewMultiParser(AllParsers)

// MultiParserResult is the result from the multi-parser
type MultiParserResult struct {
	Results []ParserResult
}

// Name returns the parser name
func (r *MultiParserResult) Name() string {
	return "multi"
}

// Lines returns the lines that this result is made from
func (r *MultiParserResult) Lines() []int {
	lines := make([]int, 0)
	for _, r := range r.Results {
		lines = append(lines, r.Lines()...)
	}
	sort.Ints(lines)
	return lines
}

// NewMultiParser returns a new MultiParser that uses all of the given parses in order of preference
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
