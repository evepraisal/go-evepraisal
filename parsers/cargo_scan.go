package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// CargoScan is the result from the cargo scan parser
type CargoScan struct {
	Items []CargoScanItem
	lines []int
}

// Name returns the parser name
func (r CargoScan) Name() string {
	return "cargo_scan"
}

// Lines returns the lines that this result is made from
func (r CargoScan) Lines() []int {
	return r.lines
}

// CargoScanItem is a single item from a cargo scan result
type CargoScanItem struct {
	Name     string
	Quantity int64
	BPC      bool
}

var reCargoScan = regexp.MustCompile(`^([\d,'\.]+) ([\S ]+)$`)

// ParseCargoScan parses cargo scans
func ParseCargoScan(input Input) (ParserResult, Input) {
	scan := &CargoScan{}
	matches, rest := regexParseLines(reCargoScan, input)
	scan.lines = regexMatchedLines(matches)

	// collect items
	matchgroup := make(map[CargoScanItem]int64)
	for _, match := range matches {
		item := CargoScanItem{Name: CleanTypeName(match[2])}

		if strings.HasSuffix(item.Name, " (Copy)") {
			item.BPC = true
			item.Name = strings.TrimSuffix(item.Name, " (Copy)")
		}

		if strings.HasSuffix(item.Name, " (Original)") {
			item.Name = strings.TrimSuffix(item.Name, " (Original)")
		}
		matchgroup[item] += ToInt(match[1])
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.Quantity = quantity
		scan.Items = append(scan.Items, item)
	}

	sort.Slice(scan.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", scan.Items[i]) < fmt.Sprintf("%v", scan.Items[j])
	})
	return scan, rest
}
