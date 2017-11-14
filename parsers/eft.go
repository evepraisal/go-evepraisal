package parsers

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

type EFT struct {
	FittingName string
	Ship        string
	Items       []ListingItem
	lines       []int
}

func (r *EFT) Name() string {
	return "eft"
}

func (r *EFT) Lines() []int {
	return r.lines
}

var reEFTHeader = regexp.MustCompile(`^\[([\S ]+), ?([\S ]+)\]$`)
var eftBlacklist = map[string]bool{
	"[empty high slot]":      true,
	"[empty low slot]":       true,
	"[empty medium slot]":    true,
	"[empty rig slot]":       true,
	"[empty subsystem slot]": true,
}

func ParseEFT(input Input) (ParserResult, Input) {
	inputLines := input.Strings()
	if len(inputLines) == 0 {
		return nil, input
	}

	line0 := inputLines[0]
	if !strings.Contains(line0, "[") || !strings.Contains(line0, "]") {
		return nil, input
	}

	headerParts := reEFTHeader.FindStringSubmatch(line0)
	if len(headerParts) == 0 {
		return nil, input
	}

	eft := &EFT{}
	eft.lines = []int{0}
	eft.Ship = headerParts[1]
	eft.FittingName = headerParts[2]

	itemsInput := StringsToInput(inputLines)
	// remove the header line (it was done this way to maintain the correct line numbers)
	delete(itemsInput, 0)

	// remove blacklisted lines
	for i, line := range itemsInput {
		_, blacklisted := eftBlacklist[strings.ToLower(line)]
		if blacklisted {
			eft.lines = append(eft.lines, i)
			delete(itemsInput, i)
		}
	}

	result, rest := ParseListing(itemsInput)
	listingResult, ok := result.(*Listing)
	if !ok {
		log.Fatal("ParseListing returned something other than parsers.Listing")
	}
	eft.Items = listingResult.Items
	eft.lines = append(eft.lines, listingResult.Lines()...)

	sort.Slice(eft.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", eft.Items[i]) < fmt.Sprintf("%v", eft.Items[j])
	})
	sort.Ints(eft.lines)
	return eft, rest
}
