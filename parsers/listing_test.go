package parsers

var listingTestCases = []Case{
	{
		"No quantity",
		`Minmatar Shuttle`,
		&Listing{
			Items: []ListingItem{
				{Name: "Minmatar Shuttle", Quantity: 1},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"quantity prefixed with x",
		`10x Minmatar Shuttle`,
		&Listing{
			Items: []ListingItem{
				{Name: "Minmatar Shuttle", Quantity: 10},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"quantity postfixed",
		`Heavy Assault Missile Launcher II 10`,
		&Listing{
			Items: []ListingItem{
				{Name: "Heavy Assault Missile Launcher II", Quantity: 10},
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
			Items: []ListingItem{{Name: "Tritanium", Quantity: 38338810}},
			lines: []int{0, 1, 2, 3},
		},
		Input{},
		false,
	}, {
		"with whitespace",
		` Tritanium
 Tritanium
Tritanium `,
		&Listing{
			Items: []ListingItem{{Name: "Tritanium", Quantity: 3}},
			lines: []int{0, 1, 2},
		},
		Input{},
		false,
	}, {
		"with capital x",
		`Tritanium x 1
Tritanium X 1`,
		&Listing{
			Items: []ListingItem{{Name: "Tritanium", Quantity: 2}},
			lines: []int{0, 1},
		},
		Input{},
		false,
	},
}
