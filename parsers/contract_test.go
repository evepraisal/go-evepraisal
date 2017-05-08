package parsers

var contractTestCases = []Case{
	{
		"Simple",
		`Rokh	1	Battleship	Ship	`,
		&Contract{
			Items: []ContractItem{ContractItem{Name: "Rokh", Quantity: 1, Type: "Battleship", Category: "Ship"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Fitted",
		`Rokh	1	Battleship	Ship	
Large Core Defense Field Extender I	1	Rig Shield	Module	Fitted`,
		&Contract{
			Items: []ContractItem{
				ContractItem{Name: "Large Core Defense Field Extender I", Quantity: 1, Type: "Rig Shield", Category: "Module", Details: "Fitted", Fitted: true},
				ContractItem{Name: "Rokh", Quantity: 1, Type: "Battleship", Category: "Ship", Details: ""},
			},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Damaged",
		`Rokh	1	Battleship	Ship	
Scorch M	1	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged
Scorch L	2	Advanced Pulse Laser Crystal	Charge	 1% damaged`,
		&Contract{
			Items: []ContractItem{
				ContractItem{Name: "Rokh", Quantity: 1, Type: "Battleship", Category: "Ship", Details: ""},
				ContractItem{Name: "Scorch L", Quantity: 2, Type: "Advanced Pulse Laser Crystal", Category: "Charge", Details: " 1% damaged"},
				ContractItem{Name: "Scorch M", Quantity: 1, Type: "Advanced Pulse Laser Crystal", Category: "Charge", Details: "Fitted 72% damaged", Fitted: true},
			},
			lines: []int{0, 1, 2},
		},
		Input{},
		true,
	}, {
		"Grouped",
		`Scorch M	1	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged
Scorch M	1	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged
Scorch M	2	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged
Scorch M	3	Advanced Pulse Laser Crystal	Charge	Fitted 72% damaged`,
		&Contract{
			Items: []ContractItem{
				ContractItem{Name: "Scorch M", Quantity: 7, Type: "Advanced Pulse Laser Crystal", Category: "Charge", Details: "Fitted 72% damaged", Fitted: true},
			},
			lines: []int{0, 1, 2, 3},
		},
		Input{},
		true,
	}, {
		"BPC",
		`Armageddon Blueprint	1	Battleship Blueprint	Blueprint	BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0`,
		&Contract{
			Items: []ContractItem{ContractItem{Name: "Armageddon Blueprint", Quantity: 1, Type: "Battleship Blueprint", Category: "Blueprint", Details: "BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Over 9000",
		`425mm Railgun I	9000	Hybrid Weapon`,
		&Contract{
			Items: []ContractItem{ContractItem{Name: "425mm Railgun I", Quantity: 9000, Type: "Hybrid Weapon"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Nothing but ship (fail)",
		`Rokh`,
		&Contract{lines: []int{}},
		Input{0: "Rokh"},
		false,
	}, {
		"Single-quote comma separator",
		`Rokh	12'000	Battleship	Ship	`,
		&Contract{
			Items: []ContractItem{ContractItem{Name: "Rokh", Quantity: 12000, Type: "Battleship", Category: "Ship"}},
			lines: []int{0},
		},
		Input{},
		true,
	},
}
