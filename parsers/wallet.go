package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Wallet is the result from the wallet parser
type Wallet struct {
	Transactions         []WalletTransaction
	ItemizedTransactions []WalletItemizedTransaction
	lines                []int
}

// Name returns the parser name
func (r *Wallet) Name() string {
	return "view_contents"
}

// Lines returns the lines that this result is made from
func (r *Wallet) Lines() []int {
	return r.lines
}

// WalletTransaction is a transaction line from a wallet log
type WalletTransaction struct {
	Datetime        string
	TransactionType string
	Amount          string
	Balance         string
	Description     string
}

// WalletItemizedTransaction is an itemized transaction line from a wallet log
type WalletItemizedTransaction struct {
	Datetime string
	Name     string
	Price    string
	Quantity int64
	Credit   string
	Currency string
	Client   string
	Location string
}

var reWallet = regexp.MustCompile(strings.Join([]string{
	`^(\d\d\d\d.\d\d.\d\d \d\d:\d\d:\d\d)\t`, // datetime
	`([\S ]+)\t`,                             // transaction type
	`([-\d,'\.]+ (?:ISK|AUR))\t`,             // amount
	`([\d,'\.]+ (?:ISK|AUR))\t`,              // balance
	`([\S ]*)$`,                              // description
}, ""))

var reWallet2 = regexp.MustCompile(strings.Join([]string{
	`^(\d\d\d\d\.\d\d\.\d\d \d\d:\d\d)\t`, // datetime
	`([\S ]+)\t`,                          // name
	`([\d,'\.]+ (?:ISK|AUR))\t`,           // price
	`([\d,'\.]+)\t`,                       // quantity
	`([-\d,'\.]+ (?:ISK|AUR))\t`,          // credit
	`(ISK|AUR)\t`,                         // currency
	`([\S ]+)\t`,                          // client
	`([\S ]+)$`,                           // location
}, ""))

// ParseWallet parses wallet text
func ParseWallet(input Input) (ParserResult, Input) {
	wallet := &Wallet{}
	matches, rest := regexParseLines(reWallet, input)
	matches2, rest := regexParseLines(reWallet2, rest)
	wallet.lines = append(wallet.lines, regexMatchedLines(matches)...)
	wallet.lines = append(wallet.lines, regexMatchedLines(matches2)...)

	for _, match := range matches {
		item := WalletTransaction{
			Datetime:        match[1],
			TransactionType: match[2],
			Amount:          match[3],
			Balance:         match[4],
			Description:     match[5],
		}
		wallet.Transactions = append(wallet.Transactions, item)
	}

	for _, match := range matches2 {
		item := WalletItemizedTransaction{
			Datetime: match[1],
			Name:     match[2],
			Price:    match[3],
			Quantity: ToInt(match[4]),
			Credit:   match[5],
			Currency: match[6],
			Client:   match[7],
			Location: match[8],
		}
		wallet.ItemizedTransactions = append(wallet.ItemizedTransactions, item)
	}

	sort.Slice(wallet.Transactions, func(i, j int) bool {
		return fmt.Sprintf("%v", wallet.Transactions[i]) < fmt.Sprintf("%v", wallet.Transactions[j])
	})
	sort.Slice(wallet.ItemizedTransactions, func(i, j int) bool {
		return fmt.Sprintf("%v", wallet.ItemizedTransactions[i]) < fmt.Sprintf("%v", wallet.ItemizedTransactions[j])
	})

	return wallet, rest
}
