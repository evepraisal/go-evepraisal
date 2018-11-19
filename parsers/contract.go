package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Contract is the result from the contract parser
type Contract struct {
	Items []ContractItem
	lines []int
}

// Name returns the parser name
func (r *Contract) Name() string {
	return "contract"
}

// Lines returns the lines that this result is made from
func (r *Contract) Lines() []int {
	return r.lines
}

// ContractItem is a single item from a contract result
type ContractItem struct {
	Name     string
	Quantity int64
	Type     string
	Category string
	Details  string
	Fitted   bool
	BPC      bool
	BPCRuns  int64
}

var reContract = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`,                 // Name
	`(` + bigNumberRegex + `*)\t`, // Quantity
	`([\S ]*)\t`,                  // type
	`([\S ]*)\t`,                  // Category
	`([\S ]*)$`,                   // Details
}, ""))

var reContractShort = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`,                 // Name
	`(` + bigNumberRegex + `*)\t`, // Quantity
	`([\S ]*)$`,                   // type
}, ""))

var reContractName = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*) (?:x|X) `,         // Name
	"(" + bigNumberRegex + "+) ", // Quantity
	`\(Item Exchange\)[\s]*`,
}, ""))

var reBPCDetails = regexp.MustCompile(`BLUEPRINT COPY(?: - Runs: ([\d]+) - )?.*`)

// ParseContract parses a contract
func ParseContract(input Input) (ParserResult, Input) {
	contract := &Contract{}
	matches, rest := regexParseLines(reContract, input)
	matches2, rest := regexParseLines(reContractShort, rest)
	matches3, rest := regexParseLines(reContractName, rest)
	contract.lines = append(regexMatchedLines(matches), regexMatchedLines(matches2)...)
	contract.lines = append(contract.lines, regexMatchedLines(matches3)...)

	// collect items
	matchgroup := make(map[ContractItem]int64)
	for _, match := range matches {
		bpc := reBPCDetails.FindStringSubmatch(match[5])
		var isBPC bool
		var bpcRuns int64
		if len(bpc) > 0 {
			isBPC = true
			if bpc[1] == "" {
				bpcRuns = 1
			} else {
				bpcRuns = ToInt(bpc[1])
			}
		}
		item := ContractItem{
			Name:     CleanTypeName(match[1]),
			Type:     match[3],
			Category: match[4],
			Details:  match[5],
			Fitted:   strings.HasPrefix(match[5], "Fitted"),
			BPC:      isBPC,
			BPCRuns:  bpcRuns,
		}

		matchgroup[item] += ToInt(match[2])
	}

	for _, match := range matches2 {
		item := ContractItem{
			Name: match[1],
			Type: match[3],
		}
		matchgroup[item] += ToInt(match[2])
	}

	for _, match := range matches3 {
		item := ContractItem{
			Name: match[1],
		}
		matchgroup[item] += ToInt(match[2])
	}

	// add items w/totals
	for item, Quantity := range matchgroup {
		item.Quantity = Quantity
		contract.Items = append(contract.Items, item)
	}

	sort.Slice(contract.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", contract.Items[i]) < fmt.Sprintf("%v", contract.Items[j])
	})
	return contract, rest
}
