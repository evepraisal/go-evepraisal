package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type SurveyScan struct {
	items []ScanItem
	lines []int
}

func (r *SurveyScan) Name() string {
	return "loot_history"
}

func (r *SurveyScan) Lines() []int {
	return r.lines
}

type ScanItem struct {
	name     string
	quantity int64
	distance string
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
		surveyScan.items = append(surveyScan.items,
			ScanItem{
				name:     match[1],
				quantity: ToInt(match[2]),
				distance: match[3],
			})
	}

	sort.Slice(surveyScan.items, func(i, j int) bool {
		return fmt.Sprintf("%v", surveyScan.items[i]) < fmt.Sprintf("%v", surveyScan.items[j])
	})
	if len(matches) > 0 {
		return surveyScan, Input{}
	}
	return surveyScan, rest
}
