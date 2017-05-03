package parsers

import (
	"regexp"
	"strings"
)

type ContractRow struct {
	name     string
	quantity int64
	_type    string
	category string
	details  string
	fitted   bool
}

var reContract = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`,   // name
	`([\d,'\.]*)\t`, // quantity
	`([\S ]*)\t`,    // type
	`([\S ]*)\t`,    // category
	`([\S ]*)$`,     // details
}, ""))

var reContractShort = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`,   // name
	`([\d,'\.]*)\t`, // quantity
	`([\S ]*)$`,     // type
}, ""))

func (r ContractRow) Name() string {
	return r.name
}

func (r ContractRow) Quantity() int64 {
	return r.quantity
}

func (r ContractRow) Volume() float64 {
	return 0
}

func ParseContract(lines []string) ([]ParserResult, []string) {
	var results []ParserResult
	matches, rest := regexParseLines(reContract, lines)
	for _, match := range matches {
		results = append(results,
			ContractRow{
				name:     match[1],
				quantity: ToInt(match[2]),
				_type:    match[3],
				category: match[4],
				details:  match[5],
			})
	}

	matches2, rest := regexParseLines(reContractShort, rest)
	for _, match := range matches2 {
		results = append(results,
			ContractRow{
				name:     match[1],
				quantity: ToInt(match[2]),
				_type:    match[3],
			})
	}

	return results, rest
}
