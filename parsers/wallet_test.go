package parsers

var walletTestCases = []Case{
	{
		"With transaction",
		`2014.01.04 05:49:31	Market Escrow	-251.00 ISK	325.22 ISK	Market escrow authorized by: Me`,
		&Wallet{
			transactions: []WalletTransaction{
				WalletTransaction{datetime: "2014.01.04 05:49:31", transactionType: "Market Escrow", amount: "-251.00 ISK", balance: "325.22 ISK", description: "Market escrow authorized by: Me"}},
			lines: []int{0}},
		Input{},
		true,
	}, {
		"With itemized transaction",
		`2014.01.04 16:08	Storm Command Center	200,000.00 ISK	1	-200,000.00 ISK	ISK	lady scarlette	Otanuomi IV - Moon 4 - Ishukone Corporation Factory`,
		&Wallet{
			itemizedTransactions: []WalletItemizedTransaction{
				WalletItemizedTransaction{datetime: "2014.01.04 16:08", name: "Storm Command Center", price: "200,000.00 ISK", quantity: 1, credit: "-200,000.00 ISK", currency: "ISK", client: "lady scarlette", location: "Otanuomi IV - Moon 4 - Ishukone Corporation Factory"}},
			lines: []int{0}},
		Input{},
		true,
	}, {
		"With itemized transaction 2",
		`2014.12.19 20:04	Medium Core Defense Capacitor Safeguard II	7'999'996.10 ISK	1	7'999'996.10 ISK	ISK	Ormand Ishikela	Jita IV - Moon 4 - Caldari Navy Assembly Plant`,
		&Wallet{
			itemizedTransactions: []WalletItemizedTransaction{
				WalletItemizedTransaction{datetime: "2014.12.19 20:04", name: "Medium Core Defense Capacitor Safeguard II", price: "7'999'996.10 ISK", quantity: 1, credit: "7'999'996.10 ISK", currency: "ISK", client: "Ormand Ishikela", location: "Jita IV - Moon 4 - Caldari Navy Assembly Plant"}},
			lines: []int{0}},
		Input{},
		true,
	},
}
