package parsers

var cargoScanTestCases = []Case{
	{
		"Simple",
		`1 Minmatar Shuttle
2 Gallente Shuttle`,
		&CargoScan{
			Items: []CargoScanItem{
				{Name: "Gallente Shuttle", Quantity: 2},
				{Name: "Minmatar Shuttle", Quantity: 1},
			},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Prefixed with new line",
		`

1 Minmatar Shuttle

`,
		&CargoScan{
			Items: []CargoScanItem{{Name: "Minmatar Shuttle", Quantity: 1}},
			lines: []int{2},
		},
		Input{0: "", 1: "", 3: "", 4: ""},
		true,
	}, {
		"BPO",
		`10 Plagioclase Mining Crystal I Blueprint (Original)`,
		&CargoScan{
			Items: []CargoScanItem{{Name: "Plagioclase Mining Crystal I Blueprint", Quantity: 10}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"BPC",
		`10 Plagioclase Mining Crystal I Blueprint (Copy)`,
		&CargoScan{
			Items: []CargoScanItem{{Name: "Plagioclase Mining Crystal I Blueprint", Quantity: 10, BPC: true}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Single quote for thousands separators",
		`12'000 Tengu`,
		&CargoScan{
			Items: []CargoScanItem{{Name: "Tengu", Quantity: 12000}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Duplicates",
		`1 Tengu
2 Tengu`,
		&CargoScan{
			Items: []CargoScanItem{{Name: "Tengu", Quantity: 3}},
			lines: []int{0, 1},
		},
		Input{},
		true,
	},
}
