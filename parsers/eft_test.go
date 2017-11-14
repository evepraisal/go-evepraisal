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
[Empty High slot]
Garde I x5`,
		&EFT{
			FittingName: "Fleet Tackle",
			Ship:        "Rifter",
			Items: []ListingItem{
				{Name: "1MN Microwarpdrive I", Quantity: 1},
				{Name: "200mm AutoCannon I", Quantity: 3},
				{Name: "EMP S", Quantity: 3},
				{Name: "Garde I", Quantity: 5},
				{Name: "Nanofiber Internal Structure I", Quantity: 2},
				{Name: "Overdrive Injector System I", Quantity: 1},
				{Name: "Stasis Webifier I", Quantity: 1},
				{Name: "Warp Disruptor I", Quantity: 1},
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
				{Name: "100MN Microwarpdrive I", Quantity: 1},
				{Name: "Aurora L", Quantity: 8},
				{Name: "Beta Reactor Control: Reaction Control I", Quantity: 1},
				{Name: "F-90 Positional Sensor Subroutines", Quantity: 1},
				{Name: "Heat Sink II", Quantity: 3},
				{Name: "Reactor Control Unit II", Quantity: 1},
				{Name: "Sensor Booster II", Quantity: 2},
				{Name: "Tachyon Beam Laser II", Quantity: 8},
				{Name: "Targeting Range Script", Quantity: 2},
				{Name: "Tracking Enhancer II", Quantity: 2}},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 14, 15, 16, 17, 18, 19, 20, 21}},
		Input{13: "", 8: ""},
		true,
	}, {
		"Basic",
		`[Rifter,test]`,
		&EFT{FittingName: "test", Ship: "Rifter", Items: nil, lines: []int{0}},
		Input{},
		true,
	}, {
		"With Ammo (Issue 30)",
		`[Phoenix, Fitting Name]

XL Torpedo Launcher I,Guristas Mjolnir XL Torpedo`,
		&EFT{
			FittingName: "Fitting Name",
			Ship:        "Phoenix",
			Items: []ListingItem{
				{Name: "Guristas Mjolnir XL Torpedo", Quantity: 1},
				{Name: "XL Torpedo Launcher I", Quantity: 1},
			},
			lines: []int{0, 2},
		},
		Input{1: ""},
		true,
	},
}
