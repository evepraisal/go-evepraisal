package parsers

var assetListTestCases = []Case{
	{
		"Simple",
		`Hurricane	1	Combat Battlecruiser`,
		[]ParserResult{AssetRow{name: "Hurricane", group: "Combat Battlecruiser", quantity: 1}},
		nil,
	}, {
		"Typical",
		`720mm Gallium Cannon	1	Projectile Weapon	Medium	High	10 m3
Damage Control II	1	Damage Control		Low	5 m3
Experimental 10MN Microwarpdrive I	1	Propulsion Module		Medium	10 m3`,
		[]ParserResult{
			AssetRow{name: "720mm Gallium Cannon", quantity: 1, group: "Projectile Weapon", category: "Medium", slot: "High", volume: 10},
			AssetRow{name: "Damage Control II", quantity: 1, group: "Damage Control", slot: "Low", volume: 5},
			AssetRow{name: "Experimental 10MN Microwarpdrive I", quantity: 1, group: "Propulsion Module", size: "Medium", volume: 10}},
		nil,
	}, {
		"Full",
		`200mm AutoCannon I	1	Projectile Weapon	Module	Small	High	5 m3	1
10MN Afterburner II	1	Propulsion Module	Module	Medium	5 m3	5	2
Warrior II	9`,
		[]ParserResult{
			AssetRow{name: "200mm AutoCannon I", quantity: 1, group: "Projectile Weapon", category: "Module", size: "Small", slot: "High", metaLevel: "1", volume: 5},
			AssetRow{name: "10MN Afterburner II", quantity: 1, group: "Propulsion Module", category: "Module", size: "Medium", metaLevel: "5", techLevel: "2", volume: 5},
			AssetRow{name: "Warrior II", quantity: 9}},
		nil,
	}, {
		"With volumes",
		`Sleeper Data Library	1080	Sleeper Components			10.82 m3`,
		[]ParserResult{AssetRow{name: "Sleeper Data Library", quantity: 1080, group: "Sleeper Components", volume: 10.82}},
		nil,
	}, {
		"With thousands separators",
		`Sleeper Data Library	1,080
Sleeper Data Library	1'080
Sleeper Data Library	1.080`,
		[]ParserResult{
			AssetRow{name: "Sleeper Data Library", quantity: 1080},
			AssetRow{name: "Sleeper Data Library", quantity: 1080},
			AssetRow{name: "Sleeper Data Library", quantity: 1080},
		},
		nil,
	},
}
