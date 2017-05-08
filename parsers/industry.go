package parsers

import (
	"fmt"
	"regexp"
	"sort"
)

type Industry struct {
	Items []IndustryItem
	lines []int
}

func (r *Industry) Name() string {
	return "industry"
}

func (r *Industry) Lines() []int {
	return r.lines
}

type IndustryItem struct {
	Name     string
	Quantity int64
}

var reIndustry = regexp.MustCompile(`^([\S ]+) \(([\d]+) Units?\)$`)

func ParseIndustry(input Input) (ParserResult, Input) {
	industry := &Industry{}
	matches, rest := regexParseLines(reIndustry, input)
	industry.lines = append(industry.lines, regexMatchedLines(matches)...)

	// collect items
	matchgroup := make(map[IndustryItem]int64)
	for _, match := range matches {
		matchgroup[IndustryItem{Name: match[1]}] += ToInt(match[2])
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.Quantity = quantity
		industry.Items = append(industry.Items, item)
	}

	sort.Slice(industry.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", industry.Items[i]) < fmt.Sprintf("%v", industry.Items[j])
	})
	return industry, rest
}
