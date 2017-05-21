package parsers

import (
	"strings"
	"testing"

	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/stretchr/testify/assert"
)

var HeuristicParserCases = []struct {
	name   string
	types  map[string]typedb.EveType
	in     string
	result ParserResult
	left   Input
}{
	{
		"example 1",
		map[string]typedb.EveType{
			"tritanium": {},
		},
		`177887021	Tritanium
44461428	Pyerite`,
		&HeuristicResult{Items: []HeuristicItem{
			{Name: "Tritanium", Quantity: 177887021}},
			lines: []int{0}},
		Input{1: "44461428\tPyerite"},
	}, {
		"example 2 - dashes",
		map[string]typedb.EveType{
			"procurer":                         {},
			"medium shield extender i":         {},
			"ice harvester ii":                 {},
			"adaptive invulnerability field i": {},
		},
		`Procurer x 1- Medium Shield Extender I x 1- Ice Harvester II x 1- Ice Harvester II x 1- Adaptive Invulnerability Field I x 1`,
		&HeuristicResult{
			Items: []HeuristicItem{
				{Name: "Adaptive Invulnerability Field I", Quantity: 1},
				{Name: "Ice Harvester II", Quantity: 2},
				{Name: "Medium Shield Extender I", Quantity: 1},
				{Name: "Procurer", Quantity: 1}},
			lines: []int{0}},
		Input{},
	},
}

func TestHeuristicParser(rt *testing.T) {
	for _, c := range HeuristicParserCases {
		rt.Run(c.name, func(t *testing.T) {
			p := HeuristicParser{
				typeDB: StaticTypeDB{c.types},
			}
			result, rest := p.Parse(StringToInput(c.in))
			assert.Equal(t, c.result, result, "results should be the same")
			assert.Equal(t, c.left, rest, "the rest should be the same")
		})
	}
}

type StaticTypeDB struct {
	typeDB map[string]typedb.EveType
}

func (db StaticTypeDB) GetType(typeName string) (typedb.EveType, bool) {
	t, ok := db.typeDB[strings.ToLower(typeName)]
	return t, ok
}
func (db StaticTypeDB) HasType(typeName string) bool {
	_, ok := db.GetType(typeName)
	return ok
}
func (db StaticTypeDB) Close() error { return nil }
