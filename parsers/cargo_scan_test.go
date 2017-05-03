package parsers

var cargoScanTestCases = []Case{
	{
		"Simple",
		`1 Minmatar Shuttle
2 Gallente Shuttle`,
		[]ParserResult{
			CargoScanRow{name: "Minmatar Shuttle", quantity: 1},
			CargoScanRow{name: "Gallente Shuttle", quantity: 2},
		},
		nil,
		true,
	}, {
		"Prefixed with new line",
		`

1 Minmatar Shuttle

`,
		[]ParserResult{CargoScanRow{name: "Minmatar Shuttle", quantity: 1}},
		[]string{"", "", "", ""},
		true,
	}, {
		"BPO",
		`10 Plagioclase Mining Crystal I Blueprint (Original)`,
		[]ParserResult{CargoScanRow{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT ORIGINAL"}},
		nil,
		true,
	}, {
		"BPC",
		`10 Plagioclase Mining Crystal I Blueprint (Copy)`,
		[]ParserResult{CargoScanRow{name: "Plagioclase Mining Crystal I Blueprint", quantity: 10, details: "BLUEPRINT COPY"}},
		nil,
		true,
	}, {
		"Single quote for thousands separators",
		`12'000 Tengu`,
		[]ParserResult{CargoScanRow{name: "Tengu", quantity: 12000}},
		nil,
		true,
	},
}
