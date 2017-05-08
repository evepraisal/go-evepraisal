package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type CargoScan struct {
	Items []CargoScanItem
	lines []int
}

func (r CargoScan) Name() string {
	return "cargo_scan"
}

func (r CargoScan) Lines() []int {
	return r.lines
}

type CargoScanItem struct {
	Name     string
	Quantity int64
	Details  string
}

var reCargoScan = regexp.MustCompile(`^([\d,'\.]+) ([\S ]+)$`)

func ParseCargoScan(input Input) (ParserResult, Input) {
	scan := &CargoScan{}
	matches, rest := regexParseLines(reCargoScan, input)
	scan.lines = regexMatchedLines(matches)
	for _, match := range matches {
		item := CargoScanItem{Name: match[2], Quantity: ToInt(match[1])}

		if strings.HasSuffix(item.Name, " (Copy)") {
			item.Details = "BLUEPRINT COPY"
			item.Name = strings.TrimSuffix(item.Name, " (Copy)")
		}

		if strings.HasSuffix(item.Name, " (Original)") {
			item.Details = "BLUEPRINT ORIGINAL"
			item.Name = strings.TrimSuffix(item.Name, " (Original)")
		}
		scan.Items = append(scan.Items, item)
	}

	sort.Slice(scan.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", scan.Items[i]) < fmt.Sprintf("%v", scan.Items[j])
	})
	return scan, rest
}
