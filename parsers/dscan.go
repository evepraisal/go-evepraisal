package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type DScan struct {
	Items []DScanItem
	lines []int
}

func (r *DScan) Name() string {
	return "dscan"
}

func (r *DScan) Lines() []int {
	return r.lines
}

type DScanItem struct {
	Name         string
	Distance     int64
	DistanceUnit string
}

var reDScan = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`, // item name
	`([\S ]*)\t`,  // Name
	`((?:([\d,'\.` + "\xc2\xa0" + `]*) (m|km|AU))|-)`, // Distance
}, ""))

func ParseDScan(input Input) (ParserResult, Input) {
	dscan := &DScan{}
	matches, rest := regexParseLines(reDScan, input)
	dscan.lines = regexMatchedLines(matches)
	for _, match := range matches {
		dscan.Items = append(dscan.Items, DScanItem{Name: CleanTypeName(match[2]), Distance: ToInt(match[4]), DistanceUnit: match[5]})
	}

	sort.Slice(dscan.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", dscan.Items[i]) < fmt.Sprintf("%v", dscan.Items[j])
	})
	return dscan, rest
}
