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
			items: []IndustryItem{
				IndustryItem{name: "Isogen", quantity: 44},
				IndustryItem{name: "Mexallon", quantity: 1027},
				IndustryItem{name: "Nocxium", quantity: 51},
				IndustryItem{name: "Pyerite", quantity: 1857},
				IndustryItem{name: "Tritanium", quantity: 4662},
			},
			lines: []int{0, 1, 2, 3, 4}},
		Input{},
		true,
	}, {
		"One unit",
		`Strontuim Clathrates (1 Unit)`,
		&Industry{
			items: []IndustryItem{IndustryItem{name: "Strontuim Clathrates", quantity: 1}},
			lines: []int{0}},
		Input{},
		true,
	},
}
