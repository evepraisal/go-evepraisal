package parsers

import (
	"fmt"
	"regexp"
	"sort"
)

type LootHistory struct {
	Items []LootItem
	lines []int
}

func (r *LootHistory) Name() string {
	return "loot_history"
}

func (r *LootHistory) Lines() []int {
	return r.lines
}

type LootItem struct {
	Time       string
	Name       string
	PlayerName string
	Quantity   int64
}

var reLootHistory = regexp.MustCompile(`(\d\d:\d\d:\d\d) ([\S ]+) has looted ([\d,'\.\ ]+) x ([\S ]+)$`)

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
