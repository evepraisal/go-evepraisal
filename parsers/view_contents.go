package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ViewContents is the result from the view contents parser
type ViewContents struct {
	Items []ViewContentsItem
	lines []int
}

// Name returns the parser name
func (r *ViewContents) Name() string {
	return "view_contents"
}

// Lines returns the lines that this result is made from
func (r *ViewContents) Lines() []int {
	return r.lines
}

// ViewContentsItem is a single item from a view contents result
type ViewContentsItem struct {
	Name     string
	Group    string
	Location string
	Quantity int64
}

var reViewContents = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`, // name
	`([\S ]*)\t`,  // group
	`((?:Cargo|Ore|Planetary Commodities) Hold|(?:Drone|Fuel|Fighter) Bay|(?:Low|Medium|High|Rig) Slot|Subsystem|Fighter Launch Tube|)\t`, // location
	`([\d,'\.]+)$`, // quantity
}, ""))

var reViewContents2 = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`,  // name
	`([\S ]*)\t`,   // group
	`([\d,'\.]+)$`, // quantity
}, ""))

// ParseViewContents parses view contents text
func ParseViewContents(input Input) (ParserResult, Input) {
	viewContents := &ViewContents{}
	matches, rest := regexParseLines(reViewContents, input)
	matches2, rest := regexParseLines(reViewContents2, rest)
	viewContents.lines = append(viewContents.lines, regexMatchedLines(matches)...)
	viewContents.lines = append(viewContents.lines, regexMatchedLines(matches2)...)

	matchgroup := make(map[ViewContentsItem]int64)
	for _, match := range matches {
		item := ViewContentsItem{
			Name:     CleanTypeName(match[1]),
			Group:    match[2],
			Location: match[3],
		}
		matchgroup[item] += ToInt(match[4])
	}

	for _, match := range matches2 {
		item := ViewContentsItem{
			Name:  CleanTypeName(match[1]),
			Group: match[2],
		}
		matchgroup[item] += ToInt(match[3])
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.Quantity = quantity
		viewContents.Items = append(viewContents.Items, item)
	}

	sort.Slice(viewContents.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", viewContents.Items[i]) < fmt.Sprintf("%v", viewContents.Items[j])
	})

	return viewContents, rest
}
