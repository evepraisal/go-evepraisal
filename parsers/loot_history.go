package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// LootHistory is the result from the loot history parser
type LootHistory struct {
	Items []LootItem
	lines []int
}

// Name returns the parser name
func (r *LootHistory) Name() string {
	return "loot_history"
}

// Lines returns the lines that this result is made from
func (r *LootHistory) Lines() []int {
	return r.lines
}

// LootItem is a single item from a loot history result
type LootItem struct {
	Time       string
	Name       string
	PlayerName string
	Quantity   int64
}

var reLootHistory = regexp.MustCompile(`(\d\d:\d\d:\d\d) ([\S ]+) has looted ([\d,'\.\ ]+) x ([\S ]+)$`)

var lootHistory2Header = "Time	Character	Item Type	Quantity	Item Group"

var reLootHistory2 = regexp.MustCompile(strings.Join([]string{
	`^(\d\d\d\d\.\d\d\.\d\d \d\d\:\d\d)`, // Datetime
	`\t([\S ]+)`,                         // Character
	`\t([\S ]+)`,                         // Item Name
	`\t([\d,'\.\ ]+)`,                    // Quantity
	`\t([\S ]+)`,                         // Item Group
}, ""))

// ParseLootHistory parses loot history text
func ParseLootHistory(input Input) (ParserResult, Input) {

	lootHistory := &LootHistory{}
	var (
		matches map[int][]string
		rest    Input
	)
	if input[0] == lootHistory2Header {
		delete(input, 0)
		matches, rest = regexParseLines(reLootHistory2, input)
		lootHistory.lines = append(lootHistory.lines, regexMatchedLines(matches)...)

		for _, match := range matches {
			lootHistory.Items = append(lootHistory.Items,
				LootItem{
					Time:       match[1],
					PlayerName: match[2],
					Name:       match[3],
					Quantity:   ToInt(match[4]),
				})
		}
	} else {
		matches, rest = regexParseLines(reLootHistory, input)
		lootHistory.lines = append(lootHistory.lines, regexMatchedLines(matches)...)

		for _, match := range matches {
			lootHistory.Items = append(lootHistory.Items,
				LootItem{
					Time:       match[1],
					PlayerName: match[2],
					Quantity:   ToInt(match[3]),
					Name:       match[4],
				})
		}
	}

	sort.Slice(lootHistory.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", lootHistory.Items[i]) < fmt.Sprintf("%v", lootHistory.Items[j])
	})
	return lootHistory, rest
}
