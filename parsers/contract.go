package parsers

import (
	"fmt"
	"regexp"
	"sort"
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

func ParseContract(input Input) (ParserResult, Input) {
	contract := &Contract{}
	matches, rest := regexParseLines(reContract, input)
	matches2, rest := regexParseLines(reContractShort, rest)
	contract.lines = append(regexMatchedLines(matches), regexMatchedLines(matches2)...)

	// collect items
	matchgroup := make(map[ContractItem]int64)
	for _, match := range matches {
		item := ContractItem{
			name:     match[1],
			_type:    match[3],
			category: match[4],
			details:  match[5],
			fitted:   strings.HasPrefix(match[5], "Fitted"),
		}

		matchgroup[item] += ToInt(match[2])
	}

	for _, match := range matches2 {
		item := ContractItem{
			name:  match[1],
			_type: match[3],
		}
		matchgroup[item] += ToInt(match[2])
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.quantity = quantity
		contract.items = append(contract.items, item)
	}

	sort.Slice(contract.items, func(i, j int) bool {
		return fmt.Sprintf("%v", contract.items[i]) < fmt.Sprintf("%v", contract.items[j])
	})
	return contract, rest
}
