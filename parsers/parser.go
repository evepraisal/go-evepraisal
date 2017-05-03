package parsers

type ParserResult interface {
	Name() string
	Quantity() int64
	Volume() float64
}

type Parser func(lines []string) ([]ParserResult, []string)

var AllParser = NewMultiParser([]Parser{
	ParseContract,
	ParseAssets,
	ParseCargoScan,
})

func NewMultiParser(parsers []Parser) Parser {
	return Parser(
		func(lines []string) ([]ParserResult, []string) {
			allResults := make([]ParserResult, 0)
			left := lines
			for _, parser := range parsers {
				var results []ParserResult
				results, left = parser(left)
				allResults = append(allResults, results...)
				if len(left) == 0 {
					break
				}
			}
			return allResults, left
		})
}
