package parsers

var piTestCases = []Case{
	{
		"Routable",
		`331.0	Aqueous Liquids	Not routed
331	Aqueous Liquids	Routed`,
		&PI{
			items: []PIItem{
				PIItem{name: "Aqueous Liquids", quantity: 331, volume: 0, routed: false},
				PIItem{name: "Aqueous Liquids", quantity: 331, volume: 0, routed: true},
			},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Quantities as floats",
		`	Aqueous Liquids	305.0	3.05`,
		&PI{
			items: []PIItem{
				PIItem{name: "Aqueous Liquids", quantity: 305.0, volume: 3.05},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Short format",
		`	Aqueous Liquids	205.0`,
		&PI{
			items: []PIItem{
				PIItem{name: "Aqueous Liquids", quantity: 205.0},
			},
			lines: []int{0},
		},
		Input{},
		true,
	},
}

//
