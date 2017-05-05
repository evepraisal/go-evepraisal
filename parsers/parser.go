package parsers

type ParserResult interface {
	Name() string
	Lines() []int
}

type Parser func(lines []string) (ParserResult, []string)

var AllParser = NewMultiParser(
	[]Parser{
		ParseContract,
		ParseAssets,
		ParseCargoScan,
		ParseDScan,
	})

// ParseAssets (type func([]string) (ParserResult, []string)) as type func([]string) ([]ParserResult, []string) in field value

type MultiParserResult struct {
	results []ParserResult
}

func (r *MultiParserResult) Name() string {
	return "multi"
}

func (r *MultiParserResult) Lines() []int {
	lines := make([]int, 0)
	for _, result := range r.results {
		lines = append(lines, result.Lines()...)
	}
	return lines
}

func NewMultiParser(parsers []Parser) Parser {
	return Parser(
		func(lines []string) (ParserResult, []string) {
			multiParserResult := &MultiParserResult{}
			left := lines
			for _, parser := range parsers {
				var result ParserResult
				result, left = parser(left)
				if len(result.Lines()) > 0 {
					multiParserResult.results = append(multiParserResult.results, result)
				}
				if len(left) == 0 {
					break
				}
			}
			return multiParserResult, left
		})
}
