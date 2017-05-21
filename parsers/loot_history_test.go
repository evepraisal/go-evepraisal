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
	},
}
