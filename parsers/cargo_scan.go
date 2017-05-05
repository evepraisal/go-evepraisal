package parsers

import (
	"regexp"
	"strings"
)

type CargoScan struct {
	items []CargoScanItem
	lines []int
}

func (r CargoScan) Name() string {
	return "cargo_scan"
}

func (r CargoScan) Lines() []int {
	return r.lines
}

type CargoScanItem struct {
	name     string
	quantity int64
	details  string
}

var reCargoScan = regexp.MustCompile(`^([\d,'\.]+) ([\S ]+)$`)

func ParseCargoScan(lines []string) (ParserResult, []string) {
	scan := &CargoScan{}
	matches, matchedLines, rest := regexParseLines(reCargoScan, lines)
	scan.lines = matchedLines
	for _, match := range matches {
		item := CargoScanItem{name: match[2], quantity: ToInt(match[1])}

		if strings.HasSuffix(item.name, " (Copy)") {
			item.details = "BLUEPRINT COPY"
			item.name = strings.TrimSuffix(item.name, " (Copy)")
		}

		if strings.HasSuffix(item.name, " (Original)") {
			item.details = "BLUEPRINT ORIGINAL"
			item.name = strings.TrimSuffix(item.name, " (Original)")
		}
		scan.items = append(scan.items, item)
	}
	return scan, rest
}
