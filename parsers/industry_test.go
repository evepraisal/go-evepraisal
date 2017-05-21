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
				{Name: "Isogen", Quantity: 44},
				{Name: "Mexallon", Quantity: 1027},
				{Name: "Nocxium", Quantity: 51},
				{Name: "Pyerite", Quantity: 1857},
				{Name: "Tritanium", Quantity: 4662},
			},
			lines: []int{0, 1, 2, 3, 4}},
		Input{},
		true,
	}, {
		"One unit",
		`Strontuim Clathrates (1 Unit)`,
		&Industry{
			Items: []IndustryItem{{Name: "Strontuim Clathrates", Quantity: 1}},
			lines: []int{0}},
		Input{},
		true,
	},
}
