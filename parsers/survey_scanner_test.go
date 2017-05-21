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
			Items: []ScanItem{
				{Name: "Pyroxeres", Quantity: 11595, Distance: "7,180 m"},
				{Name: "Pyroxeres", Quantity: 1919, Distance: "5,842 m"},
				{Name: "Pyroxeres", Quantity: 5414, Distance: "6,134 m"},
				{Name: "Veldspar", Quantity: 10, Distance: "12 km"},
				{Name: "Veldspar", Quantity: 26644, Distance: "6,115 m"},
				{Name: "Veldspar", Quantity: 26935, Distance: "12 km"},
			},
			lines: []int{0, 1, 2, 5, 6, 7},
		},
		Input{},
		true,
	},
}
