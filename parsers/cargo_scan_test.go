package parsers

var cargoScanTestCases = []Case{
	{
		"Simple",
		`1 Minmatar Shuttle
2 Gallente Shuttle`,
		&CargoScan{
			items: []CargoScanItem{
				CargoScanItem{name: "Gallente Shuttle", quantity: 2},
				CargoScanItem{name: "Minmatar Shuttle", quantity: 1},
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
			items: []CargoScanItem{CargoScanItem{name: "Minmatar Shuttle", quantity: 1}},
			lines: []int{2},
		},
		Input{0: "", 1: "", 3: "", 4: ""},
		true,
	}, {
		"BPO",
		`10 Plagioclase Mining Crystal I Blueprint (Original)`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT ORIGINAL"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"BPC",
		`10 Plagioclase Mining Crystal I Blueprint (Copy)`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT COPY"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Single quote for thousands separators",
		`12'000 Tengu`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Tengu", quantity: 12000}},
			lines: []int{0},
		},
		Input{},
		true,
	},
}
