package parsers

var miningLedgerTestCases = []Case{
	{
		"Example",
		`2018.03.01	 Bright Spodumain	24,993	399,888 m³	33,796,534 ISK	Q-02UL
2018.03.01	 Gleaming Spodumain	15,926	254,816 m³	19,282,085 ISK	7UTB-F
2018.03.02	 Gleaming Spodumain	3,393	54,288 m³	4,108,006 ISK	7UTB-F
2018.03.02	 Gneiss	48,000	240,000 m³	53,464,799 ISK	31X-RE`,
		&MiningLedger{
			Items: []MiningLedgerItem{
				{Name: "Bright Spodumain", Quantity: 24993},
				{Name: "Gleaming Spodumain", Quantity: 19319},
				{Name: "Gneiss", Quantity: 48000},
			},
			lines: []int{0, 1, 2, 3},
		},
		Input{},
		true,
	},
}
