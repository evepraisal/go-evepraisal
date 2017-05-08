package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type SurveyScan struct {
	Items []ScanItem
	lines []int
}

func (r *SurveyScan) Name() string {
	return "loot_history"
}

func (r *SurveyScan) Lines() []int {
	return r.lines
}

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

func ParseSurveyScan(input Input) (ParserResult, Input) {
	surveyScan := &SurveyScan{}
	matches, rest := regexParseLines(reSurveyScanner, input)
	surveyScan.lines = regexMatchedLines(matches)
	for _, match := range matches {
		surveyScan.Items = append(surveyScan.Items,
			ScanItem{
				Name:     match[1],
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
