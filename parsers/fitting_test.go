package parsers

var fittingTestCases = []Case{
	{
		"Basic",
		`High power
5x Heavy Missile Launcher II
Medium power
1x Large Shield Extender II
1x Dread Guristas EM Ward Amplifier
1x Domination 100MN Afterburner
1x Phased Muon Sensor Disruptor I
2x Adaptive Invulnerability Field II
Low power
Low power
1x Damage Control II
1x Reactor Control Unit II
3x Ballistic Control System II
Rig Slot
1x Medium Ancillary Current Router I
2x Medium Core Defense Field Extender I
Sub System
1x Tengu Offensive - Accelerated Ejection Bay
1x Tengu Propulsion - Fuel Catalyst
1x Tengu Defensive - Supplemental Screening
1x Tengu Electronics - Dissolution Sequencer
1x Tengu Engineering - Capacitor Regeneration Matrix
Charges
8,718x Caldari Navy Scourge Heavy Missile
1x Targeting Range Dampening Script
Drones
12 Warrior II
Fuel
Helium Isotopes`,
		&Fitting{items: []ListingItem{
			ListingItem{name: "Adaptive Invulnerability Field II", quantity: 2},
			ListingItem{name: "Ballistic Control System II", quantity: 3},
			ListingItem{name: "Caldari Navy Scourge Heavy Missile", quantity: 8718},
			ListingItem{name: "Damage Control II", quantity: 1},
			ListingItem{name: "Domination 100MN Afterburner", quantity: 1},
			ListingItem{name: "Dread Guristas EM Ward Amplifier", quantity: 1},
			ListingItem{name: "Heavy Missile Launcher II", quantity: 5},
			ListingItem{name: "Helium Isotopes", quantity: 1},
			ListingItem{name: "Large Shield Extender II", quantity: 1},
			ListingItem{name: "Medium Ancillary Current Router I", quantity: 1},
			ListingItem{name: "Medium Core Defense Field Extender I", quantity: 2},
			ListingItem{name: "Phased Muon Sensor Disruptor I", quantity: 1},
			ListingItem{name: "Reactor Control Unit II", quantity: 1},
			ListingItem{name: "Targeting Range Dampening Script", quantity: 1},
			ListingItem{name: "Tengu Defensive - Supplemental Screening", quantity: 1},
			ListingItem{name: "Tengu Electronics - Dissolution Sequencer", quantity: 1},
			ListingItem{name: "Tengu Engineering - Capacitor Regeneration Matrix", quantity: 1},
			ListingItem{name: "Tengu Offensive - Accelerated Ejection Bay", quantity: 1},
			ListingItem{name: "Tengu Propulsion - Fuel Catalyst", quantity: 1},
			ListingItem{name: "Warrior II", quantity: 12}},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28}},
		Input{},
		true,
	},
}
