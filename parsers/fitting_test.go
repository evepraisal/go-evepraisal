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
		&Fitting{
			Items: []ListingItem{
				{Name: "Adaptive Invulnerability Field II", Quantity: 2},
				{Name: "Ballistic Control System II", Quantity: 3},
				{Name: "Caldari Navy Scourge Heavy Missile", Quantity: 8718},
				{Name: "Damage Control II", Quantity: 1},
				{Name: "Domination 100MN Afterburner", Quantity: 1},
				{Name: "Dread Guristas EM Ward Amplifier", Quantity: 1},
				{Name: "Heavy Missile Launcher II", Quantity: 5},
				{Name: "Helium Isotopes", Quantity: 1},
				{Name: "Large Shield Extender II", Quantity: 1},
				{Name: "Medium Ancillary Current Router I", Quantity: 1},
				{Name: "Medium Core Defense Field Extender I", Quantity: 2},
				{Name: "Phased Muon Sensor Disruptor I", Quantity: 1},
				{Name: "Reactor Control Unit II", Quantity: 1},
				{Name: "Targeting Range Dampening Script", Quantity: 1},
				{Name: "Tengu Defensive - Supplemental Screening", Quantity: 1},
				{Name: "Tengu Electronics - Dissolution Sequencer", Quantity: 1},
				{Name: "Tengu Engineering - Capacitor Regeneration Matrix", Quantity: 1},
				{Name: "Tengu Offensive - Accelerated Ejection Bay", Quantity: 1},
				{Name: "Tengu Propulsion - Fuel Catalyst", Quantity: 1},
				{Name: "Warrior II", Quantity: 12}},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28}},
		Input{},
		true,
	},
}
