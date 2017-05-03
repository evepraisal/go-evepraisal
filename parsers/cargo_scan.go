package parsers

import (
	"regexp"
	"strings"
)

type CargoScanRow struct {
	name     string
	quantity int64
	details  string
}

func (r CargoScanRow) Name() string {
	return r.name
}

func (r CargoScanRow) Quantity() int64 {
	return r.quantity
}

func (r CargoScanRow) Volume() float64 {
	return 0
}

var reCargoScan = regexp.MustCompile(`^([\d,'\.]+) ([\S ]+)$`)

func ParseCargoScan(lines []string) ([]ParserResult, []string) {
	var results []ParserResult
	matches, rest := regexParseLines(reCargoScan, lines)
	for _, match := range matches {
		row := CargoScanRow{name: match[2], quantity: ToInt(match[1])}

		if strings.HasSuffix(row.name, " (Copy)") {
			row.details = "BLUEPRINT COPY"
			row.name = strings.TrimSuffix(row.name, " (Copy)")
		}

		if strings.HasSuffix(row.name, " (Original)") {
			row.details = "BLUEPRINT ORIGINAL"
			row.name = strings.TrimSuffix(row.name, " (Original)")
		}
		results = append(results, row)
	}
	return results, rest
}
