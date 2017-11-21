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
	}, {
		"with starting whitespace",
		` 450	125mm Railgun I
 150	Griffin
 150	Maulus
 300	Scan Resolution Dampening Script
 150	Signal Distortion Amplifier I
 150	Small Shield Extender I
 600	Stasis Webifier I
 300	Targeting Range Dampening Script
 300	Tracking Speed Disruption Script
1200	Warrior I`,
		&Listing{
			Items: []ListingItem{
				{Name: "125mm Railgun I", Quantity: 450},
				{Name: "Griffin", Quantity: 150},
				{Name: "Maulus", Quantity: 150},
				{Name: "Scan Resolution Dampening Script", Quantity: 300},
				{Name: "Signal Distortion Amplifier I", Quantity: 150},
				{Name: "Small Shield Extender I", Quantity: 150},
				{Name: "Stasis Webifier I", Quantity: 600},
				{Name: "Targeting Range Dampening Script", Quantity: 300},
				{Name: "Tracking Speed Disruption Script", Quantity: 300},
				{Name: "Warrior I", Quantity: 1200},
			},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		Input{},
		false,
	},
}
