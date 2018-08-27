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
	}, {
		"Copy Material Information",
		`Components				
Item	Required	Available	Est. Unit price	typeID
Plasma Thruster	30	0	75199.17	11530
Fernite Carbide Composite Armor Plate	294	0	14872.75	11542

Minerals				
Item	Required	Available	Est. Unit price	typeID
Morphite	38	0	10558.3	11399

Planetary materials				
Item	Required	Available	Est. Unit price	typeID
Construction Blocks	30	0	11551.94	3828

Items				
Item	Required	Available	Est. Unit price	typeID
Breacher	1	0	393441.48	598
R.A.M.- Starship Tech	3	0	314.02	11478`,
		&Industry{
			Items: []IndustryItem{
				{Name: "Breacher", Quantity: 1},
				{Name: "Construction Blocks", Quantity: 30},
				{Name: "Fernite Carbide Composite Armor Plate", Quantity: 294},
				{Name: "Morphite", Quantity: 38},
				{Name: "Plasma Thruster", Quantity: 30},
				{Name: "R.A.M.- Starship Tech", Quantity: 3},
			},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
		Input{},
		true,
	}, {
		"Copy Material Information (datacores)",
		`Datacores				
Item	Required	Available	Est. Unit price	typeID
Datacore - Mechanical Engineering	60	0	43694.05	20424
Datacore - Minmatar Starship Engineering	60	0	93418.68	20172

Optional items				
Item	Required	Available	Est. Unit price	typeID
No item selected				`,
		&Industry{
			Items: []IndustryItem{
				{Name: "Datacore - Mechanical Engineering", Quantity: 60},
				{Name: "Datacore - Minmatar Starship Engineering", Quantity: 60},
			},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7}},
		Input{},
		true,
	},
}
