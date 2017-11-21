package parsers

import (
	"fmt"
	"regexp"
	"sort"
)

type Listing struct {
	Items []ListingItem
	lines []int
}

func (r *Listing) Name() string {
	return "listing"
}

func (r *Listing) Lines() []int {
	return r.lines
}

type ListingItem struct {
	Name     string
	Quantity int64
}

var reListing = regexp.MustCompile(`^([\d,'\.]+?) ?(?:x|X)? ([\S ]+)$`)
var reListing2 = regexp.MustCompile(`^([\S ]+?) (?:x|X)? ?([\d,'\.]+)$`)
var reListing3 = regexp.MustCompile(`^([\S ]+)$`)
var reListing4 = regexp.MustCompile(`^\s*([\d,'\.]+)\t([\S ]+?)$`)
var reListingWithAmmo = regexp.MustCompile(`^([\S ]+), ?([a-zA-Z][\S ]+)$`)

func ParseListing(input Input) (ParserResult, Input) {
	listing := &Listing{}

	matchesWithAmmo, rest := regexParseLines(reListingWithAmmo, input)
	matches, rest := regexParseLines(reListing, rest)
	matches2, rest := regexParseLines(reListing2, rest)
	matches3, rest := regexParseLines(reListing3, rest)
	matches4, rest := regexParseLines(reListing4, rest)

	listing.lines = append(listing.lines, regexMatchedLines(matches)...)
	listing.lines = append(listing.lines, regexMatchedLines(matches2)...)
	listing.lines = append(listing.lines, regexMatchedLines(matches3)...)
	listing.lines = append(listing.lines, regexMatchedLines(matches4)...)
	listing.lines = append(listing.lines, regexMatchedLines(matchesWithAmmo)...)

	// collect items
	matchgroup := make(map[ListingItem]int64)
	for _, match := range matches {
		matchgroup[ListingItem{Name: CleanTypeName(match[2])}] += ToInt(match[1])
	}

	for _, match := range matches2 {
		matchgroup[ListingItem{Name: CleanTypeName(match[1])}] += ToInt(match[2])
	}

	for _, match := range matches3 {
		matchgroup[ListingItem{Name: CleanTypeName(match[1])}] += 1
	}

	for _, match := range matches4 {
		matchgroup[ListingItem{Name: CleanTypeName(match[2])}] += ToInt(match[1])
	}

	for _, match := range matchesWithAmmo {
		matchgroup[ListingItem{Name: CleanTypeName(match[1])}] += 1
		matchgroup[ListingItem{Name: CleanTypeName(match[2])}] += 1
	}

	// add items w/totals
	for item, quantity := range matchgroup {
		item.Quantity = quantity
		listing.Items = append(listing.Items, item)
	}

	sort.Slice(listing.Items, func(i, j int) bool {
		return fmt.Sprintf("%v", listing.Items[i]) < fmt.Sprintf("%v", listing.Items[j])
	})
	sort.Ints(listing.lines)
	return listing, rest
}
