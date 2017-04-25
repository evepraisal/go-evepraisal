package parsers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Case struct {
	Description  string
	Input        string
	Expected     []IResult
	ExpectedRest []string
}

var empty []string

var assetListTestCases = []Case{
	{
		"Simple",
		`Hurricane	1	Combat Battlecruiser`,
		[]IResult{AssetRow{Name: "Hurricane", Group: "Combat Battlecruiser", Quantity: 1}},
		empty,
	}, {
		"Typical",
		`720mm Gallium Cannon	1	Projectile Weapon	Medium	High	10 m3
Damage Control II	1	Damage Control		Low	5 m3
Experimental 10MN Microwarpdrive I	1	Propulsion Module		Medium	10 m3`,
		[]IResult{
			AssetRow{Name: "720mm Gallium Cannon", Quantity: 1, Group: "Projectile Weapon", Category: "Medium", Slot: "High", Volume: 10},
			AssetRow{Name: "Damage Control II", Quantity: 1, Group: "Damage Control", Slot: "Low", Volume: 5},
			AssetRow{Name: "Experimental 10MN Microwarpdrive I", Quantity: 1, Group: "Propulsion Module", Size: "Medium", Volume: 10}},
		empty,
	}, {
		"Full",
		`200mm AutoCannon I	1	Projectile Weapon	Module	Small	High	5 m3	1
10MN Afterburner II	1	Propulsion Module	Module	Medium	5 m3	5	2
Warrior II	9`,
		[]IResult{
			AssetRow{Name: "200mm AutoCannon I", Quantity: 1, Group: "Projectile Weapon", Category: "Module", Size: "Small", Slot: "High", MetaLevel: "1", Volume: 5},
			AssetRow{Name: "10MN Afterburner II", Quantity: 1, Group: "Propulsion Module", Category: "Module", Size: "Medium", MetaLevel: "5", TechLevel: "2", Volume: 5},
			AssetRow{Name: "Warrior II", Quantity: 9}},
		empty,
	}, {
		"With volumes",
		`Sleeper Data Library	1.080	Sleeper Components			10.80 m3`,
		[]IResult{AssetRow{Name: "Sleeper Data Library", Quantity: 1, Group: "Sleeper Components", Volume: 10.80}},
		empty,
	},
}

// ASSET_TABLE.add_test(u'''
// Sleeper Data Library\t1\xc2\xa0080\tSleeper Components\t\t\t10.80 m3
// ''', ([{'category': '',
//         'group': 'Sleeper Components',
//         'meta_level': None,
//         'name': 'Sleeper Data Library',
//         'quantity': 1080,
//         'size': '',
//         'slot': None,
//         'tech_level': None,
//         'volume': '10.80 m3'}], []))
// ASSET_TABLE.add_test('''
// Sleeper Data Library\t1,080\tSleeper Components\t\t\t10.80 m3
// ''', ([{'category': '',
//         'group': 'Sleeper Components',
//         'meta_level': None,
//         'name': 'Sleeper Data Library',
//         'quantity': 1080,
//         'size': '',
//         'slot': None,
//         'tech_level': None,
//         'volume': '10.80 m3'}], []))
// ASSET_TABLE.add_test('''
// Amarr Dreadnought\t1\tSpaceship Command\tSkill\t\t\t0.01 m3\t\t
// ''', ([{'category': 'Skill',
//         'slot': '',
//         'group': 'Spaceship Command',
//         'name': 'Amarr Dreadnought',
//         'volume': '0.01 m3',
//         'size': '',
//         'tech_level': '',
//         'meta_level': '',
//         'quantity': 1}], []))
// ASSET_TABLE.add_test('''
// Quafe Zero\t12\tBooster\tImplant\t\t1 \t12 m3\t\t
// ''', ([{'category': 'Implant',
//         'slot': '1 ',
//         'group': 'Booster',
//         'name': 'Quafe Zero',
//         'volume': '12 m3',
//         'size': '',
//         'tech_level': '',
//         'meta_level': '',
//         'quantity': 12}], []))
// ASSET_TABLE.add_test(
//     u"Antimatter Charge M\t100\xc2\xa0000\tHybrid Charge\tMedium\t\t"
//     u"1\xc2\xa0250 m3",
//     ([{'category': 'Medium',
//      'group': 'Hybrid Charge',
//      'meta_level': None,
//      'name': 'Antimatter Charge M',
//      'quantity': 100000,
//      'size': '',
//      'slot': None,
//      'tech_level': None,
//      'volume': '1250 m3'}], []))
// ASSET_TABLE.add_test("Hurricane\t12'000\tCombat Battlecruiser\t\t\t15,000 m3",
//                      ([{'category': '',
//                         'slot': None,
//                         'group': 'Combat Battlecruiser',
//                         'name': 'Hurricane',
//                         'volume': '15,000 m3',
//                         'size': '',
//                         'tech_level': None,
//                         'meta_level': None,
//                         'quantity': 12000}], []))

func TestParsers(rt *testing.T) {
	for _, c := range assetListTestCases {
		rt.Run(c.Description, func(t *testing.T) {
			fmt.Println(c.Input)
			result, rest := ParseAssets(strings.Split(c.Input, "\n"))
			assert.Equal(t, c.Expected, result, "results should be the same")
			assert.Equal(t, c.ExpectedRest, rest, "the rest should be the same")
		})
	}
}
