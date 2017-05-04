package parsers

import (
	"regexp"
	"strings"
)

type DScan struct {
	items []DScanItem
	raw   []string
}

func (r *DScan) Name() string {
	return "dscan"
}

func (r *DScan) Raw() string {
	return strings.Join(r.raw, "\n")
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

func ParseDScan(lines []string) (ParserResult, []string) {
	dscan := &DScan{}
	matches, raw, rest := regexParseLines(reDScan, lines)
	dscan.raw = raw
	for _, match := range matches {
		dscan.items = append(dscan.items, DScanItem{name: match[2], distance: ToInt(match[4]), distanceUnit: match[5]})
	}
	return dscan, rest
}
