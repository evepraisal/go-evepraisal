package parsers

var contractTestCases = []Case{
	{
		"Simple",
		`Rokh	1	Battleship	Ship	`,
		&Contract{
			Items: []ContractItem{{Name: "Rokh", Quantity: 1, Type: "Battleship", Category: "Ship"}},
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
				{Name: "Large Core Defense Field Extender I", Quantity: 1, Type: "Rig Shield", Category: "Module", Details: "Fitted", Fitted: true},
				{Name: "Rokh", Quantity: 1, Type: "Battleship", Category: "Ship", Details: ""},
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
				{Name: "Rokh", Quantity: 1, Type: "Battleship", Category: "Ship", Details: ""},
				{Name: "Scorch L", Quantity: 2, Type: "Advanced Pulse Laser Crystal", Category: "Charge", Details: " 1% damaged"},
				{Name: "Scorch M", Quantity: 1, Type: "Advanced Pulse Laser Crystal", Category: "Charge", Details: "Fitted 72% damaged", Fitted: true},
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
				{Name: "Scorch M", Quantity: 7, Type: "Advanced Pulse Laser Crystal", Category: "Charge", Details: "Fitted 72% damaged", Fitted: true},
			},
			lines: []int{0, 1, 2, 3},
		},
		Input{},
		true,
	}, {
		"BPC",
		`Armageddon Blueprint	1	Battleship Blueprint	Blueprint	BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0
Typhoon Blueprint	1	Battleship Blueprint	Blueprint	BLUEPRINT COPY`,
		&Contract{
			Items: []ContractItem{{
				Name:     "Armageddon Blueprint",
				Quantity: 1,
				Type:     "Battleship Blueprint",
				Category: "Blueprint",
				Details:  "BLUEPRINT COPY - Runs: 9 - Material Level: 29 - Productivity Level: 0",
				BPC:      true,
				BPCRuns:  9,
			}, {
				Name:     "Typhoon Blueprint",
				Quantity: 1,
				Type:     "Battleship Blueprint",
				Category: "Blueprint",
				Details:  "BLUEPRINT COPY",
				BPC:      true,
				BPCRuns:  1,
			}},
			lines: []int{0, 1},
		},
		Input{},
		true,
	}, {
		"Over 9000",
		`425mm Railgun I	9000	Hybrid Weapon`,
		&Contract{
			Items: []ContractItem{{Name: "425mm Railgun I", Quantity: 9000, Type: "Hybrid Weapon"}},
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
			Items: []ContractItem{{Name: "Rokh", Quantity: 12000, Type: "Battleship", Category: "Ship"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Russian with asterisks",
		`Hornet EC-300*	10	Дрон электронного противодействия*	Дрон*	Отсек для дронов
 Hornet EC-300*	10	Дрон электронного противодействия*	Дрон*	Отсек для дронов
 Praetor II*	1	Боевой дрон*	Дрон*	Отсек для дронов`,
		&Contract{
			Items: []ContractItem{
				{Name: "Hornet EC-300", Quantity: 20, Type: "Дрон электронного противодействия*", Category: "Дрон*", Details: "Отсек для дронов"},
				{Name: "Praetor II", Quantity: 1, Type: "Боевой дрон*", Category: "Дрон*", Details: "Отсек для дронов"},
			},
			lines: []int{0, 1, 2},
		},
		Input{},
		true,
	}, {
		"With spaces for separator",
		`Zydrine	10 102	Mineral	Material	`,
		&Contract{
			Items: []ContractItem{
				{Name: "Zydrine", Quantity: 10102, Type: "Mineral", Category: "Material"},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"Item exchange",
		`Zircon x 21163 (Item Exchange) `,
		&Contract{
			Items: []ContractItem{
				{Name: "Zircon", Quantity: 21163},
			},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"German thousands separator",
		`Chiral Structures*	8’865	Grundlegende Güter – Rang 1*	Planetarische Güter*	`,
		&Contract{
			Items: []ContractItem{
				{Name: "Chiral Structures", Quantity: 8865, Type: "Grundlegende Güter – Rang 1*", Category: "Planetarische Güter*"},
			},
			lines: []int{0},
		},
		Input{},
		true,
	},
}
