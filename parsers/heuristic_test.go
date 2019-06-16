package parsers

import (
	"strings"
	"testing"

	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/stretchr/testify/assert"
)

var HeuristicParserCases = []struct {
	name   string
	types  []typedb.EveType
	in     string
	result ParserResult
	left   Input
}{
	{
		"example 1",
		[]typedb.EveType{
			{Name: "Tritanium"},
		},
		`177887021	Tritanium
44461428	Pyerite`,
		&HeuristicResult{Items: []HeuristicItem{
			{Name: "Tritanium", Quantity: 177887021}},
			lines: []int{0}},
		Input{1: "44461428\tPyerite"},
	}, {
		"example 2 - dashes",
		[]typedb.EveType{
			{Name: "Procurer"},
			{Name: "Medium Shield Extender I"},
			{Name: "Ice Harvester II"},
			{Name: "Adaptive Invulnerability Field I"},
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
			db := &StaticTypeDB{
				typeNameMap: make(map[string]typedb.EveType),
				typeIDMap:   make(map[int64]typedb.EveType),
			}
			for _, t := range c.types {
				db.PutTypes([]typedb.EveType{t})
			}
			p := HeuristicParser{typeDB: db}
			result, rest := p.Parse(StringToInput(c.in))
			assert.Equal(t, c.result, result, "results should be the same")
			assert.Equal(t, c.left, rest, "the rest should be the same")
		})
	}
}

type StaticTypeDB struct {
	typeNameMap map[string]typedb.EveType
	typeIDMap   map[int64]typedb.EveType
}

func (db *StaticTypeDB) PutTypes(types []typedb.EveType) error {
	for _, t := range types {
		db.typeNameMap[strings.ToLower(t.Name)] = t
		db.typeIDMap[t.ID] = t
	}
	return nil
}

func (db *StaticTypeDB) GetType(typeName string) (typedb.EveType, bool) {
	t, ok := db.typeNameMap[strings.ToLower(typeName)]
	return t, ok
}

func (db *StaticTypeDB) HasType(typeName string) bool {
	_, ok := db.GetType(strings.ToLower(typeName))
	return ok
}

func (db *StaticTypeDB) GetTypeByID(typeID int64) (typedb.EveType, bool) {
	t, ok := db.typeIDMap[typeID]
	return t, ok
}

func (db *StaticTypeDB) ListTypes(startingTypeID int64, limit int64) ([]typedb.EveType, error) {
	return []typedb.EveType{}, nil
}

func (db *StaticTypeDB) Search(s string) []typedb.EveType {
	return nil
}

func (db *StaticTypeDB) Delete() error {
	return nil
}

func (db *StaticTypeDB) Close() error { return nil }
