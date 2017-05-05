package parsers

var cargoScanTestCases = []Case{
	{
		"Simple",
		`1 Minmatar Shuttle
2 Gallente Shuttle`,
		&CargoScan{
			items: []CargoScanItem{
				CargoScanItem{name: "Minmatar Shuttle", quantity: 1},
				CargoScanItem{name: "Gallente Shuttle", quantity: 2},
			},
			lines: []int{0, 1},
		},
		nil,
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
		[]string{"", "", "", ""},
		true,
	}, {
		"BPO",
		`10 Plagioclase Mining Crystal I Blueprint (Original)`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT ORIGINAL"}},
			lines: []int{0},
		},
		nil,
		true,
	}, {
		"BPC",
		`10 Plagioclase Mining Crystal I Blueprint (Copy)`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT COPY"}},
			lines: []int{0},
		},
		nil,
		true,
	}, {
		"Single quote for thousands separators",
		`12'000 Tengu`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Tengu", quantity: 12000}},
			lines: []int{0},
		},
		nil,
		true,
	},
}
