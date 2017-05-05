package parsers

import (
	"regexp"
	"strings"
)

type Contract struct {
	items []ContractItem
	lines []int
}

func (r *Contract) Name() string {
	return "contract"
}

func (r *Contract) Lines() []int {
	return r.lines
}

type ContractItem struct {
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

func ParseContract(lines []string) (ParserResult, []string) {
	contract := &Contract{}
	matches, matchedLines, rest := regexParseLines(reContract, lines)
	contract.lines = matchedLines
	for _, match := range matches {
		contract.items = append(contract.items,
			ContractItem{
				name:     match[1],
				quantity: ToInt(match[2]),
				_type:    match[3],
				category: match[4],
				details:  match[5],
				fitted:   strings.HasPrefix(match[5], "Fitted"),
			})
	}

	matches2, matchedLines2, rest := regexParseLines(reContractShort, rest)
	contract.lines = append(contract.lines, matchedLines2...)
	for _, match := range matches2 {
		contract.items = append(contract.items,
			ContractItem{
				name:     match[1],
				quantity: ToInt(match[2]),
				_type:    match[3],
			})
	}

	return contract, rest
}
