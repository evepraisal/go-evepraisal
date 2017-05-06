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
			items: []ViewContentsItem{
				ViewContentsItem{name: "100MN Microwarpdrive II", group: "Propulsion Module", location: "Medium Slot", quantity: 1},
				ViewContentsItem{name: "1600mm Reinforced Steel Plates II", group: "Armor Reinforcer", location: "Low Slot", quantity: 1},
				ViewContentsItem{name: "Bouncer II", group: "Combat Drone", location: "Drone Bay", quantity: 2},
				ViewContentsItem{name: "Drone Link Augmentor II", group: "Drone Control Range Module", location: "High Slot", quantity: 1},
				ViewContentsItem{name: "Giant Secure Container", group: "Secure Cargo Container", location: "", quantity: 1},
				ViewContentsItem{name: "Large Micro Jump Drive", group: "Micro Jump Drive", location: "Cargo Hold", quantity: 1},
				ViewContentsItem{name: "Large Trimark Armor Pump I", group: "Rig Armor", location: "Rig Slot", quantity: 1},
				ViewContentsItem{name: "Medium Electrochemical Capacitor Booster I", group: "Capacitor Booster", location: "Medium Slot", quantity: 1},
				ViewContentsItem{name: "Nitrogen Isotopes", group: "Ice Product", location: "Fuel Bay", quantity: 20000},
				ViewContentsItem{name: "Tengu Defensive - Adaptive Shielding", group: "Defensive Systems", location: "Subsystem", quantity: 1}},
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
			items: []ViewContentsItem{
				ViewContentsItem{name: "Festival Launcher", group: "Festival Launcher", location: "", quantity: 2},
				ViewContentsItem{name: "Hornet EC-300", group: "Electronic Warfare Drone", location: "", quantity: 50},
				ViewContentsItem{name: "Men's 'Esquire' Coat (red/gold)", group: "Outer", location: "", quantity: 1}},
			lines: []int{0, 1, 2, 3}},
		Input{},
		true,
	},
}
