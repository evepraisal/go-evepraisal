package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// SurveyScan is the result from the survey scan parser
type SurveyScan struct {
	Items []ScanItem
	lines []int
}

// Name returns the parser name
func (r *SurveyScan) Name() string {
	return "loot_history"
}

// Lines returns the lines that this result is made from
func (r *SurveyScan) Lines() []int {
	return r.lines
}

// ScanItem is a single item from a cargo scan result
type ScanItem struct {
	Name     string
	Quantity int64
	Distance string
}

var reSurveyScanner = regexp.MustCompile(strings.Join([]string{
	`^([\S ]+)\t`,          // Name
	`([\d,'\.]+)\t`,        // Quantity
	`([\d,'\.]*\ (m|km))$`, // Distance
}, ""))

// ParseSurveyScan parses text from a the survey scan screen
func ParseSurveyScan(input Input) (ParserResult, Input) {
	surveyScan := &SurveyScan{}
	matches, rest := regexParseLines(reSurveyScanner, input)
	surveyScan.lines = regexMatchedLines(matches)
	for _, match := range matches {
		surveyScan.Items = append(surveyScan.Items,
			ScanItem{
				Name:     CleanTypeName(match[1]),
				Quantity: ToInt(match[2]),
				Distance: match[3],
			})
	}

	sort.Slice(surveyScan.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", surveyScan.Items[i]) < fmt.Sprintf("%v", surveyScan.Items[j])
	})
	if len(matches) > 0 {
		return surveyScan, Input{}
	}
	return surveyScan, rest
}
