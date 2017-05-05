package parsers

import (
	"regexp"
	"sort"
	"strings"
)

type DScan struct {
	items []DScanItem
	lines []int
}

func (r *DScan) Name() string {
	return "dscan"
}

func (r *DScan) Lines() []int {
	return r.lines
}

type DScanItem struct {
	name         string
	distance     int64
	distanceUnit string
}

var reDScan = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`, // item name
	`([\S ]*)\t`,  // name
	`((?:([\d,'\.` + "\xc2\xa0" + `]*) (m|km|AU))|-)`, // distance
}, ""))

func ParseDScan(input Input) (ParserResult, Input) {
	dscan := &DScan{}
	matches, rest := regexParseLines(reDScan, input)
	dscan.lines = regexMatchedLines(matches)
	for _, match := range matches {
		dscan.items = append(dscan.items, DScanItem{name: match[2], distance: ToInt(match[4]), distanceUnit: match[5]})
	}
	sort.Slice(dscan.items, func(i, j int) bool { return dscan.items[i].name < dscan.items[j].name })
	return dscan, rest
}
