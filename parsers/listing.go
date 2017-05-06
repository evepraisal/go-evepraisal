package parsers

import (
	"fmt"
	"regexp"
	"sort"
)

type Listing struct {
	items []ListingItem
	lines []int
}

func (r *Listing) Name() string {
	return "listing"
}

func (r *Listing) Lines() []int {
	return r.lines
}

type ListingItem struct {
	name     string
	quantity int64
}

var reListing = regexp.MustCompile(`^([\d,'\.]+?) ?x? ([\S ]+)$`)
var reListing2 = regexp.MustCompile(`^([\S ]+?) x? ?([\d,'\.]+)$`)
var reListing3 = regexp.MustCompile(`^([\S ]+)$`)
var reListingWithAmmo = regexp.MustCompile(`^([\S ]+), ([\S ]+)$`)

func ParseListing(input Input) (ParserResult, Input) {
	listing := &Listing{}

	matchesWithAmmo, rest := regexParseLines(reListingWithAmmo, input)
	matches, rest := regexParseLines(reListing, rest)
	matches2, rest := regexParseLines(reListing2, rest)
	matches3, rest := regexParseLines(reListing3, rest)
	listing.lines = append(listing.lines, regexMatchedLines(matches)...)
	listing.lines = append(listing.lines, regexMatchedLines(matches2)...)
	listing.lines = append(listing.lines, regexMatchedLines(matches3)...)
	listing.lines = append(listing.lines, regexMatchedLines(matchesWithAmmo)...)

	// collect items
	matchgroup := make(map[ListingItem]int64)
	for _, match := range matches {
		matchgroup[ListingItem{name: match[2]}] += ToInt(match[1])
	}

	for _, match := range matches2 {
		matchgroup[ListingItem{name: match[1]}] += ToInt(match[2])
	}

	for _, match := range matches3 {
		matchgroup[ListingItem{name: match[1]}] += 1
	}

	for _, match := range matchesWithAmmo {
		matchgroup[ListingItem{name: match[1]}] += 1
		matchgroup[ListingItem{name: match[2]}] += 1
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.quantity = quantity
		listing.items = append(listing.items, item)
	}

	sort.Slice(listing.items, func(i, j int) bool {
		return fmt.Sprintf("%v", listing.items[i]) < fmt.Sprintf("%v", listing.items[j])
	})
	sort.Ints(listing.lines)
	return listing, rest
}
