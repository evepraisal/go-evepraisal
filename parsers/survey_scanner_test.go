package parsers

var surveyScannerTestCases = []Case{
	{
		"Basic",
		`Pyroxeres	1,919	5,842 m
Pyroxeres	11,595	7,180 m
Pyroxeres	5,414	6,134 m
Scordite
Veldspar
Veldspar	10	12 km
Veldspar	26,644	6,115 m
Veldspar	26,935	12 km`,
		&SurveyScan{
			items: []ScanItem{
				ScanItem{name: "Pyroxeres", quantity: 11595, distance: "7,180 m"},
				ScanItem{name: "Pyroxeres", quantity: 1919, distance: "5,842 m"},
				ScanItem{name: "Pyroxeres", quantity: 5414, distance: "6,134 m"},
				ScanItem{name: "Veldspar", quantity: 10, distance: "12 km"},
				ScanItem{name: "Veldspar", quantity: 26644, distance: "6,115 m"},
				ScanItem{name: "Veldspar", quantity: 26935, distance: "12 km"},
			},
			lines: []int{0, 1, 2, 5, 6, 7},
		},
		Input{},
		true,
	},
}
