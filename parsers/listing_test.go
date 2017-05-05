package parsers

var listingTestCases = []Case{
	{
		"No quantity",
		`Minmatar Shuttle`,
		&Listing{
			items: []ListingItem{
				ListingItem{name: "Minmatar Shuttle", quantity: 1},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"No quantity",
		`Minmatar Shuttle`,
		&Listing{
			items: []ListingItem{
				ListingItem{name: "Minmatar Shuttle", quantity: 1},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"quantity prefixed with x",
		`10x Minmatar Shuttle`,
		&Listing{
			items: []ListingItem{
				ListingItem{name: "Minmatar Shuttle", quantity: 10},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"quantity postfixed",
		`Heavy Assault Missile Launcher II 10`,
		&Listing{
			items: []ListingItem{
				ListingItem{name: "Heavy Assault Missile Launcher II", quantity: 10},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"with thousands separators",
		`9'584'701 x Tritanium
Tritanium 9'584'702
Tritanium x 9'584'703
9,584,704 x Tritanium`,
		&Listing{
			items: []ListingItem{ListingItem{name: "Tritanium", quantity: 38338810}},
			lines: []int{0, 1, 2, 3},
		},
		Input{},
		false,
	},
}
