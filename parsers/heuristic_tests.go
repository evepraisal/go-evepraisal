package parsers

import (
	"strings"
	"testing"

	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/stretchr/testify/assert"
)

var HeuristicParserCases = []struct {
	name   string
	in     string
	result ParserResult
	left   Input
}{
	{
		"example 1",
		`177887021	Tritanium
	44461428	Pyerite`,
		nil,
		Input{},
	},
}

func TestHeuristicParser(rt *testing.T) {
	for _, c := range HeuristicParserCases {
		rt.Run(c.name, func(t *testing.T) {
			p := HeuristicParser{
				typeDB: StaticTypeDB{map[string]typedb.EveType{
					"tritanium": typedb.EveType{},
				}},
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
