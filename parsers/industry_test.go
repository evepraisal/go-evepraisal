package parsers

var industryTestCases = []Case{
	{
		"Basic",
		`Tritanium (4662 Units)
Pyerite (1857 Units)
Mexallon (1027 Units)
Isogen (44 Units)
Nocxium (51 Units)`,
		&Industry{
			Items: []IndustryItem{
				IndustryItem{Name: "Isogen", Quantity: 44},
				IndustryItem{Name: "Mexallon", Quantity: 1027},
				IndustryItem{Name: "Nocxium", Quantity: 51},
				IndustryItem{Name: "Pyerite", Quantity: 1857},
				IndustryItem{Name: "Tritanium", Quantity: 4662},
			},
			lines: []int{0, 1, 2, 3, 4}},
		Input{},
		true,
	}, {
		"One unit",
		`Strontuim Clathrates (1 Unit)`,
		&Industry{
			Items: []IndustryItem{IndustryItem{Name: "Strontuim Clathrates", Quantity: 1}},
			lines: []int{0}},
		Input{},
		true,
	},
}
