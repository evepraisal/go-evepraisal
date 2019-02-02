package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// 2019.01.19	Corp name	miner 1	Ytterbite	8,625	86,250 m³	70,377,757 ISK
var reMoonLedgerList = regexp.MustCompile(strings.Join([]string{
	`^(\d\d\d\d\.\d\d\.\d\d)`,         // Date
	`\t([\S\ ]*)`,                     // Corp Name
	`\t([\S\ ]*)`,                     // Player Name
	`\t([\S\ ]*)`,                     // Item Name
	`\t([` + bigNumberRegex + `*)`,    // Quantity
	`\t[\S\ ]*`,                       // Volume
	`\t[` + bigNumberRegex + `* ISK$`, // CCP Price Estimate
}, ""))

var moonLedgerHeaderText = `Timestamp	Corporation	Pilot	Ore Type	Quantity	Volume	Est. Price	Ore TypeID	SolarSystemID`

// 2019.01.19	Corp Name	miner 1	Ytterbite	8625	86250	70377757	45513	30003687
var reMoonLedgerList2 = regexp.MustCompile(strings.Join([]string{
	`^(\d\d\d\d\.\d\d\.\d\d)`,      // Date
	`\t([\S\ ]*)`,                  // Corp Name
	`\t([\S\ ]*)`,                  // Player Name
	`\t([\S\ ]*)`,                  // Item Name
	`\t([` + bigNumberRegex + `*)`, // Quantity
	`\t[\S\ ]*`,                    // Volume
	`\t[` + bigNumberRegex + `*`,   // CCP Price Estimate
	`\t[\d]*`,                      // TypeID
	`\t[\d]*$`,                     // SolarSystemID
}, ""))

// MoonLedger is the result from the mining ledger parser
type MoonLedger struct {
	Items []MoonLedgerItem
	lines []int
}

// Name returns the parser name
func (r *MoonLedger) Name() string {
	return "mining_ledger"
}

// Lines returns the lines that this result is made from
func (r *MoonLedger) Lines() []int {
	return r.lines
}

// MoonLedgerItem is a single item from a mining ledger result
type MoonLedgerItem struct {
	PlayerName string
	Name       string
	Quantity   int64
}

// ParseMoonLedger will parse a mining ledger
func ParseMoonLedger(input Input) (ParserResult, Input) {
	moonLedger := &MoonLedger{}
	if len(input) > 0 && input[0] == moonLedgerHeaderText {
		moonLedger.lines = append(moonLedger.lines, 0)
		delete(input, 0)
	}

	matches, rest := regexParseLines(reMoonLedgerList, input)
	moonLedger.lines = append(moonLedger.lines, regexMatchedLines(matches)...)

	matches2, rest := regexParseLines(reMoonLedgerList2, rest)
	moonLedger.lines = append(moonLedger.lines, regexMatchedLines(matches2)...)

	matchgroup := make(map[MoonLedgerItem]int64)
	for _, match := range matches {
		item := MoonLedgerItem{
			PlayerName: match[3],
			Name:       CleanTypeName(match[4]),
		}
		matchgroup[item] += ToInt(match[5])
	}

	for _, match := range matches2 {
		item := MoonLedgerItem{
			PlayerName: match[3],
			Name:       CleanTypeName(match[4]),
		}
		matchgroup[item] += ToInt(match[5])
	}

	for item, quantity := range matchgroup {
		item.Quantity = quantity
		moonLedger.Items = append(moonLedger.Items, item)
	}

	sort.Slice(moonLedger.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", moonLedger.Items[i]) < fmt.Sprintf("%v", moonLedger.Items[j])
	})
	return moonLedger, rest
}
