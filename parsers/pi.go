package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type PI struct {
	items []PIItem
	lines []int
}

func (r *PI) Name() string {
	return "pi"
}

func (r *PI) Lines() []int {
	return r.lines
}

type PIItem struct {
	name     string
	quantity int64
	volume   float64
	routed   bool
}

var rePI1 = regexp.MustCompile(strings.Join([]string{
	`^([\d,'\.]+)\t`,          // quantity
	`([\S ]+)\t`,              // name
	`((Routed|Not\ routed))$`, // routed
}, ""))

var rePI2 = regexp.MustCompile(strings.Join([]string{
	`^\t`,           // icon (ignore)
	`([\S ]+)\t`,    // name
	`([\d,'\.]+)\t`, // quantity
	`([\d,'\.]+)$`,  // volume
}, ""))

var rePI3 = regexp.MustCompile(strings.Join([]string{
	`^\t`,          // icon (ignore)
	`([\S ]+)\t`,   // name
	`([\d,'\.]+)$`, // quantity
}, ""))

func ParsePI(input Input) (ParserResult, Input) {
	pi := &PI{}
	matches1, rest := regexParseLines(rePI1, input)
	matches2, rest := regexParseLines(rePI2, rest)
	matches3, rest := regexParseLines(rePI3, rest)
	pi.lines = append(pi.lines, regexMatchedLines(matches1)...)
	pi.lines = append(pi.lines, regexMatchedLines(matches2)...)
	pi.lines = append(pi.lines, regexMatchedLines(matches3)...)

	// collect items
	matchgroup := make(map[PIItem]int64)
	for _, match := range matches1 {
		item := PIItem{name: match[2], routed: match[3] == "Routed"}

		matchgroup[item] += int64(ToFloat64(match[1]))
	}

	for _, match := range matches2 {
		item := PIItem{name: match[1], volume: ToFloat64(match[3])}
		matchgroup[item] += int64(ToFloat64(match[2]))
	}

	for _, match := range matches3 {
		item := PIItem{name: match[1]}
		matchgroup[item] += int64(ToFloat64(match[2]))
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.quantity = quantity
		pi.items = append(pi.items, item)
	}

	sort.Slice(pi.items, func(i, j int) bool {
		return fmt.Sprintf("%v", pi.items[i]) < fmt.Sprintf("%v", pi.items[j])
	})
	return pi, rest
}
