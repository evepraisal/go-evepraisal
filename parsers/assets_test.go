package parsers

var assetListTestCases = []Case{
	{
		"Simple",
		`Hurricane	1	Combat Battlecruiser`,
		&AssetList{
			items: []AssetItem{AssetItem{name: "Hurricane", group: "Combat Battlecruiser", quantity: 1}},
			lines: []int{0},
		},
		Input{},
		false, // This clashes with the simple contract format
	}, {
		"Typical",
		`720mm Gallium Cannon	1	Projectile Weapon	Medium	High	10 m3
Damage Control II	1	Damage Control		Low	5 m3
Experimental 10MN Microwarpdrive I	1	Propulsion Module		Medium	10 m3`,
		&AssetList{
			items: []AssetItem{
				AssetItem{name: "720mm Gallium Cannon", quantity: 1, group: "Projectile Weapon", category: "Medium", slot: "High", volume: 10},
				AssetItem{name: "Damage Control II", quantity: 1, group: "Damage Control", slot: "Low", volume: 5},
				AssetItem{name: "Experimental 10MN Microwarpdrive I", quantity: 1, group: "Propulsion Module", size: "Medium", volume: 10},
			},
			lines: []int{0, 1, 2}},
		Input{},
		true,
	}, {
		"Full",
		`200mm AutoCannon I	1	Projectile Weapon	Module	Small	High	5 m3	1
10MN Afterburner II	1	Propulsion Module	Module	Medium	5 m3	5	2
Warrior II	9`,
		&AssetList{
			items: []AssetItem{
				AssetItem{name: "10MN Afterburner II", quantity: 1, group: "Propulsion Module", category: "Module", size: "Medium", metaLevel: "5", techLevel: "2", volume: 5},
				AssetItem{name: "200mm AutoCannon I", quantity: 1, group: "Projectile Weapon", category: "Module", size: "Small", slot: "High", metaLevel: "1", volume: 5},
				AssetItem{name: "Warrior II", quantity: 9},
			},
			lines: []int{0, 1, 2},
		},
		Input{},
		true,
	}, {
		"With volumes",
		`Sleeper Data Library	1080	Sleeper Components			10.82 m3`,
		&AssetList{
			items: []AssetItem{AssetItem{name: "Sleeper Data Library", quantity: 1080, group: "Sleeper Components", volume: 10.82}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"With thousands separators",
		`Sleeper Data Library	1,080
Sleeper Data Library	1'080
Sleeper Data Library	1.080`,
		&AssetList{
			items: []AssetItem{
				AssetItem{name: "Sleeper Data Library", quantity: 1080},
				AssetItem{name: "Sleeper Data Library", quantity: 1080},
				AssetItem{name: "Sleeper Data Library", quantity: 1080},
			},
			lines: []int{0, 1, 2},
		},
		Input{},
		false,
	},
}
