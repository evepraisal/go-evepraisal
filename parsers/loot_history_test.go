package parsers

var lootHistoryTestCases = []Case{
	{
		"Simple",
		`03:21:19 Some dude has looted 5 x Garde II`,
		&LootHistory{
			Items: []LootItem{{Time: "03:21:19", PlayerName: "Some dude", Quantity: 5, Name: "Garde II"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Duplicate entry except for time",
		`03:21:19 Some dude has looted 5 x Garde II
04:22:20 Some dude has looted 5 x Garde II`,
		&LootHistory{
			Items: []LootItem{
				{Time: "03:21:19", PlayerName: "Some dude", Quantity: 5, Name: "Garde II"},
				{Time: "04:22:20", PlayerName: "Some dude", Quantity: 5, Name: "Garde II"},
			},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Thousands Separator",
		`03:21:19 A cool dude has looted 5'000 x Garde II`,
		&LootHistory{
			Items: []LootItem{{Time: "03:21:19", PlayerName: "A cool dude", Quantity: 5000, Name: "Garde II"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Alternative number format",
		`17:07:32 Nathan Ohmiras has looted 34 016 x Viscous Pyroxeres`,
		&LootHistory{
			Items: []LootItem{{Time: "17:07:32", PlayerName: "Nathan Ohmiras", Quantity: 34016, Name: "Viscous Pyroxeres"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"issue 75",
		`Time	Character	Item Type	Quantity	Item Group
2018.06.23 00:19	Kado Vargadana	5MN Quad LiF Restrained Microwarpdrive	1	Propulsion Module	
2018.06.23 00:19	Kado Vargadana	Faint Epsilon Scoped Warp Scrambler	1	Warp Scrambler	
2018.06.23 00:19	Kado Vargadana	X5 Enduring Stasis Webifier	1	Stasis Web	
2018.06.23 00:19	Kado Vargadana	'Refuge' Adaptive Nano Plating I	1	Armor Coating	
2018.06.23 00:19	Kado Vargadana	Anode Light Electron Particle Cannon I	1	Hybrid Weapon	`,
		&LootHistory{
			Items: []LootItem{
				{Time: "2018.06.23 00:19", Name: "'Refuge' Adaptive Nano Plating I", PlayerName: "Kado Vargadana", Quantity: 1},
				{Time: "2018.06.23 00:19", Name: "5MN Quad LiF Restrained Microwarpdrive", PlayerName: "Kado Vargadana", Quantity: 1},
				{Time: "2018.06.23 00:19", Name: "Anode Light Electron Particle Cannon I", PlayerName: "Kado Vargadana", Quantity: 1},
				{Time: "2018.06.23 00:19", Name: "Faint Epsilon Scoped Warp Scrambler", PlayerName: "Kado Vargadana", Quantity: 1},
				{Time: "2018.06.23 00:19", Name: "X5 Enduring Stasis Webifier", PlayerName: "Kado Vargadana", Quantity: 1}},
			lines: []int{1, 2, 3, 4, 5},
		},
		Input{},
		true,
	},
}
