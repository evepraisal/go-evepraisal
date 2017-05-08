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
			FittingName: "Fleet Tackle",
			Ship:        "Rifter",
			Items: []ListingItem{
				ListingItem{Name: "1MN Microwarpdrive I", Quantity: 1},
				ListingItem{Name: "200mm AutoCannon I", Quantity: 3},
				ListingItem{Name: "EMP S", Quantity: 3},
				ListingItem{Name: "Garde I", Quantity: 5},
				ListingItem{Name: "Nanofiber Internal Structure I", Quantity: 2},
				ListingItem{Name: "Overdrive Injector System I", Quantity: 1},
				ListingItem{Name: "Stasis Webifier I", Quantity: 1},
				ListingItem{Name: "Warp Disruptor I", Quantity: 1},
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
			FittingName: "Pimpin' Sniper Fit",
			Ship:        "Apocalypse",
			Items: []ListingItem{
				ListingItem{Name: "100MN Microwarpdrive I", Quantity: 1},
				ListingItem{Name: "Aurora L", Quantity: 8},
				ListingItem{Name: "Beta Reactor Control: Reaction Control I", Quantity: 1},
				ListingItem{Name: "F-90 Positional Sensor Subroutines", Quantity: 1},
				ListingItem{Name: "Heat Sink II", Quantity: 3},
				ListingItem{Name: "Reactor Control Unit II", Quantity: 1},
				ListingItem{Name: "Sensor Booster II", Quantity: 2},
				ListingItem{Name: "Tachyon Beam Laser II", Quantity: 8},
				ListingItem{Name: "Targeting Range Script", Quantity: 2},
				ListingItem{Name: "Tracking Enhancer II", Quantity: 2}},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 14, 15, 16, 17, 18, 19, 20, 21}},
		Input{13: "", 8: ""},
		true,
	}, {
		"Basic",
		`[Rifter,test]`,
		&EFT{FittingName: "test", Ship: "Rifter", Items: nil, lines: []int{0}},
		Input{},
		true,
	},
}
