package parsers

var piTestCases = []Case{
	{
		"Routable",
		`331.0	Aqueous Liquids	Not routed
331	Aqueous Liquids	Routed`,
		&PI{
			Items: []PIItem{
				{Name: "Aqueous Liquids", Quantity: 331, Volume: 0, Routed: false},
				{Name: "Aqueous Liquids", Quantity: 331, Volume: 0, Routed: true},
			},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Quantities as floats",
		`	Aqueous Liquids	305.0	3.05`,
		&PI{
			Items: []PIItem{
				{Name: "Aqueous Liquids", Quantity: 305.0, Volume: 3.05},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Short format",
		`	Aqueous Liquids	205.0`,
		&PI{
			Items: []PIItem{
				{Name: "Aqueous Liquids", Quantity: 205.0},
			},
			lines: []int{0},
		},
		Input{},
		true,
	},
}

//
