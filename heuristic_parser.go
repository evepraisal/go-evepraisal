package evepraisal

import (
	"log"
	"strings"

	"github.com/evepraisal/go-evepraisal/parsers"
)

const (
	HEURISTIC_ITEM     = iota
	HEURISTIC_QUANTITY = iota
	HEURISTIC_IGNORE   = iota
)

var HeuristicSpecs = [][]int{
	{HEURISTIC_IGNORE, HEURISTIC_ITEM, HEURISTIC_IGNORE, HEURISTIC_QUANTITY},
	{HEURISTIC_QUANTITY, HEURISTIC_IGNORE, HEURISTIC_ITEM},
	{HEURISTIC_ITEM, HEURISTIC_QUANTITY},
	{HEURISTIC_QUANTITY, HEURISTIC_ITEM},
	{HEURISTIC_IGNORE, HEURISTIC_ITEM},
	{HEURISTIC_ITEM},
}

type HeuristicParser struct {
	TypeDB TypeDB
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
		if s != "" {
			newParts = append(newParts, s)
		}
	}
	return newParts
}

func (p *HeuristicParser) Parse(input parsers.Input) (parsers.ParserResult, parsers.Input) {
	rest := make(parsers.Input)
	for i, line := range input {
		parts := removeEmpty(heuristicTrimStrings(strings.Split(line, "\t"), ", _=-[]"))
		if len(parts) == 1 {
			parts = removeEmpty(heuristicTrimStrings(strings.Split(line, "  "), ", _=-[]"))
		}
		if len(parts) == 1 {
			parts = removeEmpty(heuristicTrimStrings(strings.Split(line, " "), ", _=-[]"))
		}

		if len(parts) == 1 {
			rest[i] = line
			continue
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
				case HEURISTIC_IGNORE:
				case HEURISTIC_ITEM:
					name = parts[index]
					if !p.TypeDB.HasType(name) {
						matched = false
						break
					}
				case HEURISTIC_QUANTITY:
					quantity = parsers.ToInt(parts[index])
					if quantity == 0 {
						matched = false
						break
					}
				}
			}

			if matched {
				log.Println("FOUND", name, quantity)
				break
			}
		}
	}

	return nil, rest
}
