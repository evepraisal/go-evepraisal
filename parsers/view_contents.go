package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type ViewContents struct {
	items []ViewContentsItem
	lines []int
}

func (r *ViewContents) Name() string {
	return "view_contents"
}

func (r *ViewContents) Lines() []int {
	return r.lines
}

type ViewContentsItem struct {
	name     string
	group    string
	location string
	quantity int64
}

var reViewContents = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`, // name
	`([\S ]*)\t`,  // group
	`(Cargo Hold|(?:Drone|Fuel) Bay|(?:Low|Medium|High|Rig) Slot|Subsystem|)\t`, // location
	`([\d,'\.]+)$`, // quantity
}, ""))

var reViewContents2 = regexp.MustCompile(strings.Join([]string{
	`^([\S ]*)\t`,  // name
	`([\S ]*)\t`,   // group
	`([\d,'\.]+)$`, // quantity
}, ""))

func ParseViewContents(input Input) (ParserResult, Input) {
	viewContents := &ViewContents{}
	matches, rest := regexParseLines(reViewContents, input)
	matches2, rest := regexParseLines(reViewContents2, rest)
	viewContents.lines = append(viewContents.lines, regexMatchedLines(matches)...)
	viewContents.lines = append(viewContents.lines, regexMatchedLines(matches2)...)

	matchgroup := make(map[ViewContentsItem]int64)
	for _, match := range matches {
		item := ViewContentsItem{
			name:     match[1],
			group:    match[2],
			location: match[3],
		}
		matchgroup[item] += ToInt(match[4])
	}

	for _, match := range matches2 {
		item := ViewContentsItem{
			name:  match[1],
			group: match[2],
		}
		matchgroup[item] += ToInt(match[3])
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.quantity = quantity
		viewContents.items = append(viewContents.items, item)
	}

	sort.Slice(viewContents.items, func(i, j int) bool {
		return fmt.Sprintf("%v", viewContents.items[i]) < fmt.Sprintf("%v", viewContents.items[j])
	})

	return viewContents, rest
}
