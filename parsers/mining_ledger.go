package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// 2018.03.01	 Bright Spodumain	24,993	399,888 m³	33,796,534 ISK	Q-02UL
var reMiningLedgerList = regexp.MustCompile(strings.Join([]string{
	`^(\d\d\d\d\.\d\d\.\d\d)`,        // Date
	`\t([\S\ ]*)`,                    // Name
	`\t([` + bigNumberRegex + `*)`,   // Quantity
	`\t[\S\ ]*`,                      // Volume
	`\t[` + bigNumberRegex + `* ISK`, // CCP Price Estimate
	`\t[\S\ ]*$`,                     // System
}, ""))

// MiningLedger is the result from the mining ledger parser
type MiningLedger struct {
	Items []MiningLedgerItem
	lines []int
}

// Name returns the parser name
func (r *MiningLedger) Name() string {
	return "mining_ledger"
}

// Lines returns the lines that this result is made from
func (r *MiningLedger) Lines() []int {
	return r.lines
}

// MiningLedgerItem is a single item from a mining ledger result
type MiningLedgerItem struct {
	Name     string
	Quantity int64
}

// ParseMiningLedger will parse a mining ledger
func ParseMiningLedger(input Input) (ParserResult, Input) {
	miningLedger := &MiningLedger{}
	matches, rest := regexParseLines(reMiningLedgerList, input)
	miningLedger.lines = regexMatchedLines(matches)
	matchgroup := make(map[MiningLedgerItem]int64)
	for _, match := range matches {
		matchgroup[MiningLedgerItem{Name: CleanTypeName(match[2])}] += ToInt(match[3])
	}

	for item, quantity := range matchgroup {
		item.Quantity = quantity
		miningLedger.Items = append(miningLedger.Items, item)
	}

	sort.Slice(miningLedger.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", miningLedger.Items[i]) < fmt.Sprintf("%v", miningLedger.Items[j])
	})
	return miningLedger, rest
}
