package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Compare is the result from the asset parser
type Compare struct {
	Items []CompareItem
	lines []int
}

// Name returns the parser name
func (r *Compare) Name() string {
	return "compare"
}

// Lines returns the lines that this result is made from
func (r *Compare) Lines() []int {
	return r.lines
}

// CompareItem is a single item parsed from an asset list
type CompareItem struct {
	Name string
}

var reCompare = regexp.MustCompile(strings.Join([]string{
	`^([\S\ ]*)`, // Name
	`\t(Tech I|Tech II|Tech III|Faction|Deadspace|Storyline)`, // Size
	`[\S\ \t]*`, // Ignore rest
}, ""))

// ParseCompare will parse an compare window
func ParseCompare(input Input) (ParserResult, Input) {
	compareResult := &Compare{}
	matches, rest := regexParseLines(reCompare, input)
	compareResult.lines = regexMatchedLines(matches)
	for _, match := range matches {
		compareResult.Items = append(compareResult.Items,
			CompareItem{
				Name: CleanTypeName(match[1]),
			})
	}
	sort.Slice(compareResult.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", compareResult.Items[i]) < fmt.Sprintf("%v", compareResult.Items[j])
	})
	return compareResult, rest
}
