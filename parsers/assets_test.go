package parsers

var assetListTestCases = []Case{
	{
		"Simple",
		`Hurricane	1	Combat Battlecruiser`,
		&AssetList{
			Items: []AssetItem{{Name: "Hurricane", Group: "Combat Battlecruiser", Quantity: 1}},
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
			Items: []AssetItem{
				{Name: "720mm Gallium Cannon", Quantity: 1, Group: "Projectile Weapon", Category: "Medium", Slot: "High", Volume: 10},
				{Name: "Damage Control II", Quantity: 1, Group: "Damage Control", Slot: "Low", Volume: 5},
				{Name: "Experimental 10MN Microwarpdrive I", Quantity: 1, Group: "Propulsion Module", Size: "Medium", Volume: 10},
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
			Items: []AssetItem{
				{Name: "10MN Afterburner II", Quantity: 1, Group: "Propulsion Module", Category: "Module", Size: "Medium", MetaLevel: "5", TechLevel: "2", Volume: 5},
				{Name: "200mm AutoCannon I", Quantity: 1, Group: "Projectile Weapon", Category: "Module", Size: "Small", Slot: "High", MetaLevel: "1", Volume: 5},
				{Name: "Warrior II", Quantity: 9},
			},
			lines: []int{0, 1, 2},
		},
		Input{},
		true,
	}, {
		"With Volumes",
		`Sleeper Data Library	1080	Sleeper Components			10.82 m3`,
		&AssetList{
			Items: []AssetItem{{Name: "Sleeper Data Library", Quantity: 1080, Group: "Sleeper Components", Volume: 10.82}},
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
			Items: []AssetItem{
				{Name: "Sleeper Data Library", Quantity: 1080},
				{Name: "Sleeper Data Library", Quantity: 1080},
				{Name: "Sleeper Data Library", Quantity: 1080},
			},
			lines: []int{0, 1, 2},
		},
		Input{},
		false,
	}, {
		"With empty quantity",
		`Sleeper Data Library	`,
		&AssetList{
			Items: []AssetItem{
				{Name: "Sleeper Data Library", Quantity: 1},
			},
			lines: []int{0},
		},
		Input{},
		false,
	}, {
		"With asterisk",
		`Armor Plates*	477	Geborgene Materialien*`,
		&AssetList{
			Items: []AssetItem{
				{Name: "Armor Plates", Quantity: 477, Group: "Geborgene Materialien*"},
			},
			lines: []int{0},
		},
		Input{},
		false,
	}, {
		"With spaces in numbers",
		`Robotics	741	Specialized Commodities			4 446 m3	76 705 081,83 ISK`,
		&AssetList{
			Items: []AssetItem{
				{Name: "Robotics", Quantity: 741, Group: "Specialized Commodities", PriceEstimate: 76705081.83, Volume: 4446},
			},
			lines: []int{0},
		},
		Input{},
		false,
	}, {
		"With 'spaces' in numbers",
		`Guardian Angels 'Advanced' Cerebral Accelerator	1	Booster		10	1 m3	37 805 997.92 ISK`,
		&AssetList{
			Items: []AssetItem{
				{Name: "Guardian Angels 'Advanced' Cerebral Accelerator", Quantity: 1, Group: "Booster", Slot: "10", Volume: 1.0, PriceEstimate: 37805997.92},
			},
			lines: []int{0},
		},
		Input{},
		false,
	}, {
		"With 'spaces' in numbers 2",
		"Mexallon\t1\u00a0667\u00a0487\tMineral\t\t\t16\u00a0674,87 m3\t128\u00a0696\u00a0646,66 ISK",
		&AssetList{
			Items: []AssetItem{
				{Name: "Mexallon", Quantity: 1667487, Group: "Mineral", Volume: 16674.87, PriceEstimate: 128696646.66},
			},
			lines: []int{0},
		},
		Input{},
		false,
	}, {
		"m^3",
		"Evaporite Deposits	1 452	Moon Materials			72,60 м^3	7 533 164,76 ISK",
		&AssetList{
			Items: []AssetItem{
				{Name: "Evaporite Deposits", Quantity: 1452, Group: "Moon Materials", Volume: 72.60, PriceEstimate: 7533164.76},
			},
			lines: []int{0},
		},
		Input{},
		false,
	},
}
