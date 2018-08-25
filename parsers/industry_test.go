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
	}, {
		"Blueprint Tab",
		`Cap Booster 3200 Blueprint	10	0	-1	2	NU4-2G - Writer's Workshop	Item hangar	Capacitor Booster Charge
Deflection Shield Emitter Blueprint	10	20	-1	0	NU4-2G - Writer's Workshop	Item hangar	Construction Components
Victorieux Luxury Yacht Blueprint	0	0	1	Cruiser
2 x Medium Warhead Rigor Catalyst I Blueprint	0	0	-1	3	NU4-2G - Writer's Workshop	Item hangar	Rig Launcher`,
		&Industry{
			Items: []IndustryItem{
				{Name: "Cap Booster 3200 Blueprint", Quantity: 1, BPC: true, BPCRuns: 2},
				{Name: "Deflection Shield Emitter Blueprint", Quantity: 1},
				{Name: "Medium Warhead Rigor Catalyst I Blueprint", Quantity: 2, BPC: true, BPCRuns: 3},
				{Name: "Victorieux Luxury Yacht Blueprint", Quantity: 1, BPC: true, BPCRuns: 1},
			},
			lines: []int{0, 1, 2, 3}},
		Input{},
		true,
	},
}
