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
			raw: []string{"1 Minmatar Shuttle", "2 Gallente Shuttle"},
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
			raw:   []string{"1 Minmatar Shuttle"},
		},
		[]string{"", "", "", ""},
		true,
	}, {
		"BPO",
		`10 Plagioclase Mining Crystal I Blueprint (Original)`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT ORIGINAL"}},
			raw:   []string{"10 Plagioclase Mining Crystal I Blueprint (Original)"},
		},
		nil,
		true,
	}, {
		"BPC",
		`10 Plagioclase Mining Crystal I Blueprint (Copy)`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT COPY"}},
			raw:   []string{"10 Plagioclase Mining Crystal I Blueprint (Copy)"},
		},
		nil,
		true,
	}, {
		"Single quote for thousands separators",
		`12'000 Tengu`,
		&CargoScan{
			items: []CargoScanItem{CargoScanItem{name: "Tengu", quantity: 12000}},
			raw:   []string{"12'000 Tengu"},
		},
		nil,
		true,
	},
}
