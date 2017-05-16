package parsers

import (
	"fmt"
	"sort"

	"github.com/evepraisal/go-evepraisal/typedb"
)

type ContextListingParser struct {
	typeDB typedb.TypeDB
}

func NewContextListingParser(typeDB typedb.TypeDB) Parser {
	p := &ContextListingParser{typeDB: typeDB}
	return p.Parse
}

func (p *ContextListingParser) Parse(input Input) (ParserResult, Input) {
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
	for i, match := range matches {
		name := CleanString(match[2])
		if p.typeDB.HasType(name) {
			matchgroup[ListingItem{Name: name}] += ToInt(match[1])
		} else {
			rest[i] = input[i]
		}
	}

	for i, match := range matches2 {
		name := CleanString(match[1])
		if p.typeDB.HasType(name) {
			matchgroup[ListingItem{Name: name}] += ToInt(match[2])
		} else {
			rest[i] = input[i]
		}
	}

	for i, match := range matches3 {
		name := CleanString(match[1])
		if p.typeDB.HasType(name) {
			matchgroup[ListingItem{Name: name}] += 1
		} else {
			rest[i] = input[i]
		}
	}

	for i, match := range matchesWithAmmo {
		name1 := CleanString(match[1])
		name2 := CleanString(match[2])
		if p.typeDB.HasType(name1) && p.typeDB.HasType(name2) {
			matchgroup[ListingItem{Name: name1}] += 1
			matchgroup[ListingItem{Name: name2}] += 1
		} else {
			rest[i] = input[i]
		}
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
