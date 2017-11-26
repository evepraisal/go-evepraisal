package parsers

import (
	"fmt"
	"regexp"
	"sort"
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

// ParseLootHistory parses loot history text
func ParseLootHistory(input Input) (ParserResult, Input) {
	lootHistory := &LootHistory{}
	matches, rest := regexParseLines(reLootHistory, input)
	lootHistory.lines = regexMatchedLines(matches)
	for _, match := range matches {
		lootHistory.Items = append(lootHistory.Items,
			LootItem{
				Time:       match[1],
				PlayerName: match[2],
				Quantity:   ToInt(match[3]),
				Name:       match[4],
			})
	}

	sort.Slice(lootHistory.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", lootHistory.Items[i]) < fmt.Sprintf("%v", lootHistory.Items[j])
	})
	return lootHistory, rest
}
