package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// DScan is the result from the cargo scan parser
type DScan struct {
	Items []DScanItem
	lines []int
}

// Name returns the parser name
func (r *DScan) Name() string {
	return "dscan"
}

// Lines returns the lines that this result is made from
func (r *DScan) Lines() []int {
	return r.lines
}

// DScanItem is a single item from a dscan result
type DScanItem struct {
	Name         string
	Distance     float64
	DistanceUnit string
}

var reDScan = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`, // item name
	`([\S ]*)\t`,  // Name
	`((?:([\d,'\.` + "\xc2\xa0" + `]*) (m|km|AU))|-)`, // Distance
}, ""))

// ParseDScan parses a d-scan
func ParseDScan(input Input) (ParserResult, Input) {
	dscan := &DScan{}
	matches, rest := regexParseLines(reDScan, input)
	dscan.lines = regexMatchedLines(matches)
	for _, match := range matches {
		dscan.Items = append(dscan.Items, DScanItem{Name: CleanTypeName(match[2]), Distance: ToFloat64(match[4]), DistanceUnit: match[5]})
	}

	sort.Slice(dscan.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", dscan.Items[i]) < fmt.Sprintf("%v", dscan.Items[j])
	})
	return dscan, rest
}
