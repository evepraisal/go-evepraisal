package parsers

var contractTestCases = []Case{
	{
		"Simple",
		`Rokh	1	Battleship	Ship	`,
		&Contract{
			items: []ContractItem{ContractItem{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", fitted: false}},
			raw:   []string{"Rokh\t1\tBattleship\tShip\t"},
		},
		nil,
		true,
	}, {
		"Fitted",
		`Rokh	1	Battleship	Ship	
Large Core Defense Field Extender I	1	Rig Shield	Module	Fitted`,
		&Contract{
			items: []ContractItem{
				ContractItem{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", details: "", fitted: false},
				ContractItem{name: "Large Core Defense Field Extender I", quantity: 1, _type: "Rig Shield", category: "Module", details: "Fitted", fitted: false}},
			raw: []string{"Rokh\t1\tBattleship\tShip\t", "Large Core Defense Field Extender I\t1\tRig Shield\tModule\tFitted"},
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
				ContractItem{name: "Rokh", quantity: 1, _type: "Battleship", category: "Ship", details: "", fitted: false},
				ContractItem{name: "Scorch M", quantity: 1, _type: "Advanced Pulse Laser Crystal", category: "Charge", details: "Fitted 72% damaged", fitted: false},
				ContractItem{name: "Scorch L", quantity: 2, _type: "Advanced Pulse Laser Crystal", category: "Charge", details: " 1% damaged", fitted: false}},
			raw: []string{"Rokh\t1\tBattleship\tShip\t", "Scorch M\t1\tAdvanced Pulse Laser Crystal\tCharge\tFitted 72% damaged", "Scorch L\t2\tAdvanced Pulse Laser Crystal\tCharge\t 1% damaged"},
		},
		nil,
		true,
	}, {
		"BPC",
		`Armageddon Blueprint	1	Battleship Blueprint	Blueprint	BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0`,
		&Contract{
			items: []ContractItem{ContractItem{name: "Armageddon Blueprint", quantity: 1, _type: "Battleship Blueprint", category: "Blueprint", details: "BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0", fitted: false}},
			raw:   []string{"Armageddon Blueprint\t1\tBattleship Blueprint\tBlueprint\tBLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0"},
		},
		nil,
		true,
	}, {
		"Over 9000",
		`425mm Railgun I	9000	Hybrid Weapon`,
		&Contract{
			items: []ContractItem{ContractItem{name: "425mm Railgun I", quantity: 9000, _type: "Hybrid Weapon"}},
			raw:   []string{"425mm Railgun I\t9000\tHybrid Weapon"},
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
			raw:   []string{"Rokh\t12'000\tBattleship\tShip\t"},
		},
		nil,
		true,
	},
}
