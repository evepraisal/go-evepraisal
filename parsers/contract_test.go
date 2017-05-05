package parsers

var contractTestCases = []Case{
	{
		"Simple",
		`Rokh	1	Battleship	Ship	`,
		&Contract{
			items: []ContractItem{ContractItem{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship"}},
			lines: []int{0},
		},
		nil,
		true,
	}, {
		"Fitted",
		`Rokh	1	Battleship	Ship	
Large Core Defense Field Extender I	1	Rig Shield	Module	Fitted`,
		&Contract{
			items: []ContractItem{
				ContractItem{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", details: ""},
				ContractItem{name: "Large Core Defense Field Extender I", quantity: 1, _type: "Rig Shield", category: "Module", details: "Fitted", fitted: true},
			},
			lines: []int{0, 1},
		},
		nil,
		true,
	}, {
		"Damaged",
		`Rokh	1	Battleship	Ship	
Scorch M	1	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged
Scorch L	2	Advanced Pulse Laser Crystal	Charge	 1% damaged`,
		&Contract{
			items: []ContractItem{
				ContractItem{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", details: ""},
				ContractItem{name: "Scorch M", quantity: 1, _type: "Advanced Pulse Laser Crystal", category: "Charge", details: "Fitted 72% damaged", fitted: true},
				ContractItem{name: "Scorch L", quantity: 2, _type: "Advanced Pulse Laser Crystal", category: "Charge", details: " 1% damaged"},
			},
			lines: []int{0, 1, 2},
		},
		nil,
		true,
	}, {
		"BPC",
		`Armageddon Blueprint	1	Battleship Blueprint	Blueprint	BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0`,
		&Contract{
			items: []ContractItem{ContractItem{name: "Armageddon Blueprint", quantity: 1, _type: "Battleship Blueprint", category: "Blueprint", details: "BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0"}},
			lines: []int{0},
		},
		nil,
		true,
	}, {
		"Over 9000",
		`425mm Railgun I	9000	Hybrid Weapon`,
		&Contract{
			items: []ContractItem{ContractItem{name: "425mm Railgun I", quantity: 9000, _type: "Hybrid Weapon"}},
			lines: []int{0},
		},
		nil,
		true,
	}, {
		"Nothing but ship (fail)",
		`Rokh`,
		&Contract{},
		[]string{"Rokh"},
		false,
	}, {
		"Single-quote comma separator",
		`Rokh	12'000	Battleship	Ship	`,
		&Contract{
			items: []ContractItem{ContractItem{name: "Rokh", quantity: 12000, _type: "Battleship", category: "Ship"}},
			lines: []int{0},
		},
		nil,
		true,
	},
}
