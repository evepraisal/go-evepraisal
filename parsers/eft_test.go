package parsers

var eftTestCases = []Case{
	{
		"Basic",
		`[Rifter, Fleet Tackle]
Nanofiber Internal Structure I
Nanofiber Internal Structure I
Overdrive Injector System I
Stasis Webifier I
Warp Disruptor I
1MN Microwarpdrive I
200mm AutoCannon I, EMP S
200mm AutoCannon I, EMP S
200mm AutoCannon I, EMP S
[empty high slot]
[empty high slot]
Garde I x5`,
		&EFT{
			name: "Fleet Tackle",
			ship: "Rifter",
			items: []ListingItem{
				ListingItem{name: "1MN Microwarpdrive I", quantity: 1},
				ListingItem{name: "200mm AutoCannon I", quantity: 3},
				ListingItem{name: "EMP S", quantity: 3},
				ListingItem{name: "Garde I", quantity: 5},
				ListingItem{name: "Nanofiber Internal Structure I", quantity: 2},
				ListingItem{name: "Overdrive Injector System I", quantity: 1},
				ListingItem{name: "Stasis Webifier I", quantity: 1},
				ListingItem{name: "Warp Disruptor I", quantity: 1},
			},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		Input{},
		true,
	}, {
		"Basic",
		`[Apocalypse, Pimpin' Sniper Fit]
Heat Sink II
Heat Sink II
Heat Sink II
Tracking Enhancer II
Tracking Enhancer II
Reactor Control Unit II
Beta Reactor Control: Reaction Control I

100MN Microwarpdrive I
Sensor Booster II, Targeting Range Script
Sensor Booster II, Targeting Range Script
F-90 Positional Sensor Subroutines

Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L
Tachyon Beam Laser II, Aurora L`,
		&EFT{
			name: "Pimpin' Sniper Fit",
			ship: "Apocalypse",
			items: []ListingItem{
				ListingItem{name: "100MN Microwarpdrive I", quantity: 1},
				ListingItem{name: "Aurora L", quantity: 8},
				ListingItem{name: "Beta Reactor Control: Reaction Control I", quantity: 1},
				ListingItem{name: "F-90 Positional Sensor Subroutines", quantity: 1},
				ListingItem{name: "Heat Sink II", quantity: 3},
				ListingItem{name: "Reactor Control Unit II", quantity: 1},
				ListingItem{name: "Sensor Booster II", quantity: 2},
				ListingItem{name: "Tachyon Beam Laser II", quantity: 8},
				ListingItem{name: "Targeting Range Script", quantity: 2},
				ListingItem{name: "Tracking Enhancer II", quantity: 2}},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 14, 15, 16, 17, 18, 19, 20, 21}},
		Input{13: "", 8: ""},
		true,
	}, {
		"Basic",
		`[Rifter,test]`,
		&EFT{name: "test", ship: "Rifter", items: nil, lines: []int{0}},
		Input{},
		true,
	},
}
