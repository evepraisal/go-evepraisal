package evepraisal

import (
	"testing"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/stretchr/testify/assert"
)

var HeuristicParserCases = []struct {
	name   string
	in     string
	result parsers.ParserResult
	left   parsers.Input
}{
	{
		"example 1",
		`177887021	Tritanium
	44461428	Pyerite`,
		nil,
		parsers.Input{},
	},
}

func TestHeuristicParser(rt *testing.T) {
	for _, c := range HeuristicParserCases {
		rt.Run(c.name, func(t *testing.T) {
			p := HeuristicParser{
				TypeMap: map[string]MarketType{
					"tritanium": MarketType{},
				},
			}
			result, rest := p.Parse(parsers.StringToInput(c.in))
			assert.Equal(t, c.result, result, "results should be the same")
			assert.Equal(t, c.left, rest, "the rest should be the same")
		})
	}
}
