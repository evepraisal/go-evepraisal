package parsers

import (
	"bytes"
	"strings"

	"github.com/evepraisal/go-evepraisal/typedb"
)

const (
	sHeuristicItem     = iota
	sHeuristicQuantity = iota
	sHeuristicIgnore   = iota
)

var HeuristicSpecs = [][]int{
	{sHeuristicIgnore, sHeuristicItem, sHeuristicIgnore, sHeuristicQuantity},
	{sHeuristicQuantity, sHeuristicIgnore, sHeuristicItem},
	{sHeuristicItem, sHeuristicQuantity},
	{sHeuristicQuantity, sHeuristicItem},
}

type HeuristicParser struct {
	typeDB typedb.TypeDB
}

func heuristicTrimStrings(parts []string, trim string) []string {
	for i := range parts {
		parts[i] = strings.Trim(parts[i], trim)
	}
	return parts
}

func removeEmpty(parts []string) []string {
	newParts := make([]string, 0)
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s != "" {
			newParts = append(newParts, s)
		}
	}
	return newParts
}

type HeuristicResult struct {
	Items []HeuristicItem
	lines []int
}

func (r *HeuristicResult) Name() string {
	return "heuristic"
}

func (r *HeuristicResult) Lines() []int {
	return r.lines
}

type HeuristicItem struct {
	Name     string
	Quantity int64
}

func NewHeuristicParser(typeDB typedb.TypeDB) Parser {
	p := &HeuristicParser{typeDB: typeDB}
	return p.Parse
}

func (p *HeuristicParser) Parse(input Input) (ParserResult, Input) {
	var items []HeuristicItem
	var lines []int
	rest := make(Input)
	for i, line := range input {
		var lineResults []HeuristicItem
		lineResults = p.heuristicMethod1(line)
		if lineResults != nil {
			items = append(items, lineResults...)
			lines = append(lines, i)
			continue
		}

		lineResults = p.heuristicMethod2(line)
		if lineResults != nil {
			items = append(items, lineResults...)
			lines = append(lines, i)
			continue
		}

		// We give up. :(
		rest[i] = line
	}

	return &HeuristicResult{
		Items: items,
		lines: lines,
	}, rest
}

func (p *HeuristicParser) heuristicMethod1(line string) []HeuristicItem {
	// Let's try tab separators
	parts := removeEmpty(heuristicTrimStrings(strings.Split(line, "\t"), ", _=-[]*"))
	if len(parts) == 1 {
		// Let's try double space separators
		parts = removeEmpty(heuristicTrimStrings(strings.Split(line, "  "), ", _=-[]*"))
	}
	if len(parts) == 1 {
		// Let's try dash separators
		parts = removeEmpty(heuristicTrimStrings(strings.Split(line, "-"), ", _=-[]*"))
	}
	if len(parts) == 1 {
		// Let's try space separators
		parts = removeEmpty(heuristicTrimStrings(strings.Split(line, " "), ", _=-[]*"))
	}

	if len(parts) == 1 {
		return nil
	}

	for _, spec := range HeuristicSpecs {
		if len(parts) < len(spec) {
			continue
		}

		name := ""
		quantity := int64(1)
		matched := true
		for index, specPart := range spec {
			switch specPart {
			case sHeuristicIgnore:
			case sHeuristicItem:
				name = parts[index]
				if !p.typeDB.HasType(name) {
					matched = false
					break
				}
			case sHeuristicQuantity:
				quantity = ToInt(parts[index])
				if quantity == 0 {
					matched = false
					break
				}
			}
		}

		if matched {
			return []HeuristicItem{{Name: name, Quantity: quantity}}
		}
	}

	r, _ := NewContextListingParser(p.typeDB)(StringsToInput(parts))
	if len(r.Lines()) > 0 {
		switch r := r.(type) {
		case *Listing:
			var items []HeuristicItem
			for _, item := range r.Items {
				items = append(items, HeuristicItem{Name: item.Name, Quantity: item.Quantity})
			}
			return items
		}
	}
	return nil
}

func (p *HeuristicParser) heuristicMethod2(line string) []HeuristicItem {
	var b bytes.Buffer
	for _, part := range strings.Fields(line) {
		b.WriteString(strings.Trim(part, ",\t "))
		name := b.String()
		if p.typeDB.HasType(name) {
			return []HeuristicItem{{Name: name, Quantity: 1}}
		}
	}
	return nil
}
