package parsers

var moonLedgerTestCases = []Case{
	{
		"Example",
		`2019.01.19	Corp name	miner 1	Ytterbite	8,625	86,250 m³	70,377,757 ISK
2019.01.19	Corp name	miner 1	Bountiful Ytterbite	2,940	29,400 m³	38,004,556 ISK
2019.01.19	Corp name	miner 2	Ytterbite	7,667	76,670 m³	62,560,726 ISK
2019.01.19	Corp name	miner 2	Bountiful Ytterbite	612	6,120 m³	7,911,152 ISK`,
		&MoonLedger{
			Items: []MoonLedgerItem{
				{PlayerName: "miner 1", Name: "Bountiful Ytterbite", Quantity: 2940},
				{PlayerName: "miner 1", Name: "Ytterbite", Quantity: 8625},
				{PlayerName: "miner 2", Name: "Bountiful Ytterbite", Quantity: 612},
				{PlayerName: "miner 2", Name: "Ytterbite", Quantity: 7667},
			},
			lines: []int{0, 1, 2, 3}},
		Input{},
		true,
	}, {
		"Example using copy to clipboard button",
		`Timestamp	Corporation	Pilot	Ore Type	Quantity	Volume	Est. Price	Ore TypeID	SolarSystemID
2019.01.19	Corp Name	miner 1	Ytterbite	8625	86250	70377757	45513	30003687
2019.01.19	Corp Name	miner 1	Bountiful Ytterbite	2940	29400.0	38004556	46318	30003687
2019.01.19	Corp Name	miner 2	Ytterbite	7667	76670	62560726	45513	30003687
2019.01.19	Corp Name	miner 2	Bountiful Ytterbite	612	6120.0	7911152	46318	30003687`,
		&MoonLedger{
			Items: []MoonLedgerItem{
				{PlayerName: "miner 1", Name: "Bountiful Ytterbite", Quantity: 2940},
				{PlayerName: "miner 1", Name: "Ytterbite", Quantity: 8625},
				{PlayerName: "miner 2", Name: "Bountiful Ytterbite", Quantity: 612},
				{PlayerName: "miner 2", Name: "Ytterbite", Quantity: 7667},
			},
			lines: []int{0, 1, 2, 3, 4}},
		Input{},
		true,
	},
}
