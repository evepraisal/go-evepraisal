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
				ViewContentsItem{Name: "100MN Microwarpdrive II", Group: "Propulsion Module", Location: "Medium Slot", Quantity: 1},
				ViewContentsItem{Name: "1600mm Reinforced Steel Plates II", Group: "Armor Reinforcer", Location: "Low Slot", Quantity: 1},
				ViewContentsItem{Name: "Bouncer II", Group: "Combat Drone", Location: "Drone Bay", Quantity: 2},
				ViewContentsItem{Name: "Drone Link Augmentor II", Group: "Drone Control Range Module", Location: "High Slot", Quantity: 1},
				ViewContentsItem{Name: "Giant Secure Container", Group: "Secure Cargo Container", Location: "", Quantity: 1},
				ViewContentsItem{Name: "Large Micro Jump Drive", Group: "Micro Jump Drive", Location: "Cargo Hold", Quantity: 1},
				ViewContentsItem{Name: "Large Trimark Armor Pump I", Group: "Rig Armor", Location: "Rig Slot", Quantity: 1},
				ViewContentsItem{Name: "Medium Electrochemical Capacitor Booster I", Group: "Capacitor Booster", Location: "Medium Slot", Quantity: 1},
				ViewContentsItem{Name: "Nitrogen Isotopes", Group: "Ice Product", Location: "Fuel Bay", Quantity: 20000},
				ViewContentsItem{Name: "Tengu Defensive - Adaptive Shielding", Group: "Defensive Systems", Location: "Subsystem", Quantity: 1}},
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
				ViewContentsItem{Name: "Festival Launcher", Group: "Festival Launcher", Location: "", Quantity: 2},
				ViewContentsItem{Name: "Hornet EC-300", Group: "Electronic Warfare Drone", Location: "", Quantity: 50},
				ViewContentsItem{Name: "Men's 'Esquire' Coat (red/gold)", Group: "Outer", Location: "", Quantity: 1}},
			lines: []int{0, 1, 2, 3}},
		Input{},
		true,
	},
}
