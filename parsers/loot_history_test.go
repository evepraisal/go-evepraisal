package parsers

var lootHistoryTestCases = []Case{
	{
		"Simple",
		`03:21:19 Some dude has looted 5 x Garde II`,
		&LootHistory{
			items: []LootItem{LootItem{time: "03:21:19", playerName: "Some dude", quantity: 5, name: "Garde II"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Duplicate entry except for time",
		`03:21:19 Some dude has looted 5 x Garde II
04:22:20 Some dude has looted 5 x Garde II`,
		&LootHistory{
			items: []LootItem{
				LootItem{time: "03:21:19", playerName: "Some dude", quantity: 5, name: "Garde II"},
				LootItem{time: "04:22:20", playerName: "Some dude", quantity: 5, name: "Garde II"},
			},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Thousands Separator",
		`03:21:19 A cool dude has looted 5'000 x Garde II`,
		&LootHistory{
			items: []LootItem{LootItem{time: "03:21:19", playerName: "A cool dude", quantity: 5000, name: "Garde II"}},
			lines: []int{0},
		},
		Input{},
		true,
	},
}
