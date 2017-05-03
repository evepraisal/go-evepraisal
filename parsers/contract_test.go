package parsers

var contractTestCases = []Case{
	{
		"Simple",
		`Rokh	1	Battleship	Ship	`,
		[]ParserResult{ContractRow{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", fitted: false}},
		nil,
		true,
	}, {
		"Fitted",
		`Rokh	1	Battleship	Ship	
Large Core Defense Field Extender I	1	Rig Shield	Module	Fitted`,
		[]ParserResult{
			ContractRow{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", details: "", fitted: false},
			ContractRow{name: "Large Core Defense Field Extender I", quantity: 1, _type: "Rig Shield", category: "Module", details: "Fitted", fitted: false}},
		nil,
		true,
	}, {
		"Damaged",
		`Rokh	1	Battleship	Ship	
Scorch M	1	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged
Scorch L	2	Advanced Pulse Laser Crystal	Charge	 1% damaged`,
		[]ParserResult{
			ContractRow{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", details: "", fitted: false},
			ContractRow{name: "Scorch M", quantity: 1, _type: "Advanced Pulse Laser Crystal", category: "Charge", details: "Fitted 72% damaged", fitted: false},
			ContractRow{name: "Scorch L", quantity: 2, _type: "Advanced Pulse Laser Crystal", category: "Charge", details: " 1% damaged", fitted: false}},
		nil,
		true,
	}, {
		"BPC",
		`Armageddon Blueprint	1	Battleship Blueprint	Blueprint	BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0`,
		[]ParserResult{ContractRow{name: "Armageddon Blueprint", quantity: 1, _type: "Battleship Blueprint", category: "Blueprint", details: "BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0", fitted: false}},
		nil,
		true,
	}, {
		"Over 9000",
		`425mm Railgun I	9000	Hybrid Weapon`,
		[]ParserResult{ContractRow{name: "425mm Railgun I", quantity: 9000, _type: "Hybrid Weapon"}},
		nil,
		true,
	}, {
		"Nothing but ship (fail)",
		`Rokh`,
		nil,
		[]string{"Rokh"},
		false,
	}, {
		"Single-quote comma separator",
		`Rokh	12'000	Battleship	Ship	`,
		[]ParserResult{ContractRow{name: "Rokh", quantity: 12000, _type: "Battleship", category: "Ship"}},
		nil,
		true,
	},
}
