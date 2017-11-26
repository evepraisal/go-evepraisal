package parsers

var walletTestCases = []Case{
	{
		"With transaction",
		`2014.01.04 05:49:31	Market Escrow	-251.00 ISK	325.22 ISK	Market escrow authorized by: Me`,
		&Wallet{
			Transactions: []WalletTransaction{
				{
					Datetime:        "2014.01.04 05:49:31",
					TransactionType: "Market Escrow",
					Amount:          "-251.00 ISK",
					Balance:         "325.22 ISK",
					Description:     "Market escrow authorized by: Me"}},
			lines: []int{0}},
		Input{},
		true,
	}, {
		"With itemized transaction",
		`2014.01.04 16:08	Storm Command Center	200,000.00 ISK	1	-200,000.00 ISK	ISK	lady scarlette	Otanuomi IV - Moon 4 - Ishukone Corporation Factory`,
		&Wallet{
			ItemizedTransactions: []WalletItemizedTransaction{
				{
					Datetime: "2014.01.04 16:08",
					Name:     "Storm Command Center",
					Price:    "200,000.00 ISK",
					Quantity: 1,
					Credit:   "-200,000.00 ISK",
					Currency: "ISK",
					Client:   "lady scarlette",
					Location: "Otanuomi IV - Moon 4 - Ishukone Corporation Factory",
				}},
			lines: []int{0}},
		Input{},
		true,
	}, {
		"With itemized transaction 2",
		`2014.12.19 20:04	Medium Core Defense Capacitor Safeguard II	7'999'996.10 ISK	1	7'999'996.10 ISK	ISK	Ormand Ishikela	Jita IV - Moon 4 - Caldari Navy Assembly Plant`,
		&Wallet{
			ItemizedTransactions: []WalletItemizedTransaction{
				{
					Datetime: "2014.12.19 20:04",
					Name:     "Medium Core Defense Capacitor Safeguard II",
					Price:    "7'999'996.10 ISK",
					Quantity: 1,
					Credit:   "7'999'996.10 ISK",
					Currency: "ISK",
					Client:   "Ormand Ishikela",
					Location: "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
				}},
			lines: []int{0}},
		Input{},
		true,
	},
}
