package parsers

var viewContentsTestCases = []Case{
	{
		"Routable",
		`1600mm Reinforced Steel Plates II	Armor Reinforcer	Low Slot	1
100MN Microwarpdrive II	Propulsion Module	Medium Slot	1
Bouncer II	Combat Drone	Drone Bay	1
Bouncer II	Combat Drone	Drone Bay	1
Nitrogen Isotopes	Ice Product	Fuel Bay	20000
Drone Link Augmentor II	Drone Control Range Module	High Slot	1
Large Micro Jump Drive	Micro Jump Drive	Cargo Hold	1
Tengu Defensive - Adaptive Shielding	Defensive Systems	Subsystem	1
Large Trimark Armor Pump I	Rig Armor	Rig Slot	1
Medium Electrochemical Capacitor Booster I	Capacitor Booster	Medium Slot	1
Giant Secure Container	Secure Cargo Container		1`,
		&ViewContents{
			Items: []ViewContentsItem{
				{Name: "100MN Microwarpdrive II", Group: "Propulsion Module", Location: "Medium Slot", Quantity: 1},
				{Name: "1600mm Reinforced Steel Plates II", Group: "Armor Reinforcer", Location: "Low Slot", Quantity: 1},
				{Name: "Bouncer II", Group: "Combat Drone", Location: "Drone Bay", Quantity: 2},
				{Name: "Drone Link Augmentor II", Group: "Drone Control Range Module", Location: "High Slot", Quantity: 1},
				{Name: "Giant Secure Container", Group: "Secure Cargo Container", Location: "", Quantity: 1},
				{Name: "Large Micro Jump Drive", Group: "Micro Jump Drive", Location: "Cargo Hold", Quantity: 1},
				{Name: "Large Trimark Armor Pump I", Group: "Rig Armor", Location: "Rig Slot", Quantity: 1},
				{Name: "Medium Electrochemical Capacitor Booster I", Group: "Capacitor Booster", Location: "Medium Slot", Quantity: 1},
				{Name: "Nitrogen Isotopes", Group: "Ice Product", Location: "Fuel Bay", Quantity: 20000},
				{Name: "Tengu Defensive - Adaptive Shielding", Group: "Defensive Systems", Location: "Subsystem", Quantity: 1}},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		Input{},
		true,
	}, {
		"Routable",
		`Festival Launcher	Festival Launcher	1
Festival Launcher	Festival Launcher	1
Hornet EC-300	Electronic Warfare Drone	50
Men's 'Esquire' Coat (red/gold)	Outer	1`,
		&ViewContents{
			Items: []ViewContentsItem{
				{Name: "Festival Launcher", Group: "Festival Launcher", Location: "", Quantity: 2},
				{Name: "Hornet EC-300", Group: "Electronic Warfare Drone", Location: "", Quantity: 50},
				{Name: "Men's 'Esquire' Coat (red/gold)", Group: "Outer", Location: "", Quantity: 1}},
			lines: []int{0, 1, 2, 3}},
		Input{},
		true,
	}, {
		"Ore Hold - Issue #28",
		`Compressed Vivid Hemorphite	Hemorphite	Ore Hold	51`,
		&ViewContents{
			Items: []ViewContentsItem{
				{Name: "Compressed Vivid Hemorphite", Group: "Hemorphite", Location: "Ore Hold", Quantity: 51},
			},
			lines: []int{0}},
		Input{},
		true,
	}, {
		"Fighters - Issue #94",
		`Einherji II	Light Fighter	Fighter Bay	27`,
		&ViewContents{
			Items: []ViewContentsItem{
				{Name: "Einherji II", Group: "Light Fighter", Location: "Fighter Bay", Quantity: 27},
			},
			lines: []int{0}},
		Input{},
		true,
	}, {
		"Planetray Commodieties Hold",
		`Transmitter	Refined Commodities - Tier 2	Planetary Commodities Hold	5605`,
		&ViewContents{
			Items: []ViewContentsItem{
				{Name: "Transmitter", Group: "Refined Commodities - Tier 2", Location: "Planetary Commodities Hold", Quantity: 5605},
			},
			lines: []int{0}},
		Input{},
		true,
	},
}
