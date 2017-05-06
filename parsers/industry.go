package parsers

import (
	"fmt"
	"regexp"
	"sort"
)

type Industry struct {
	items []IndustryItem
	lines []int
}

func (r *Industry) Name() string {
	return "industry"
}

func (r *Industry) Lines() []int {
	return r.lines
}

type IndustryItem struct {
	name     string
	quantity int64
}

var reIndustry = regexp.MustCompile(`^([\S ]+) \(([\d]+) Units?\)$`)

func ParseIndustry(input Input) (ParserResult, Input) {
	industry := &Industry{}
	matches, rest := regexParseLines(reIndustry, input)
	industry.lines = append(industry.lines, regexMatchedLines(matches)...)

	// collect items
	matchgroup := make(map[IndustryItem]int64)
	for _, match := range matches {
		matchgroup[IndustryItem{name: match[1]}] += ToInt(match[2])
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.quantity = quantity
		industry.items = append(industry.items, item)
	}

	sort.Slice(industry.items, func(i, j int) bool {
		return fmt.Sprintf("%v", industry.items[i]) < fmt.Sprintf("%v", industry.items[j])
	})
	return industry, rest
}
