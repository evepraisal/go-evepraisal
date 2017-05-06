package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Wallet struct {
	transactions         []WalletTransaction
	itemizedTransactions []WalletItemizedTransaction
	lines                []int
}

func (r *Wallet) Name() string {
	return "view_contents"
}

func (r *Wallet) Lines() []int {
	return r.lines
}

type WalletTransaction struct {
	datetime        string
	transactionType string
	amount          string
	balance         string
	description     string
}

type WalletItemizedTransaction struct {
	datetime string
	name     string
	price    string
	quantity int64
	credit   string
	currency string
	client   string
	location string
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

func ParseWallet(input Input) (ParserResult, Input) {
	wallet := &Wallet{}
	matches, rest := regexParseLines(reWallet, input)
	matches2, rest := regexParseLines(reWallet2, rest)
	wallet.lines = append(wallet.lines, regexMatchedLines(matches)...)
	wallet.lines = append(wallet.lines, regexMatchedLines(matches2)...)

	for _, match := range matches {
		item := WalletTransaction{
			datetime:        match[1],
			transactionType: match[2],
			amount:          match[3],
			balance:         match[4],
			description:     match[5],
		}
		wallet.transactions = append(wallet.transactions, item)
	}

	for _, match := range matches2 {
		item := WalletItemizedTransaction{
			datetime: match[1],
			name:     match[2],
			price:    match[3],
			quantity: ToInt(match[4]),
			credit:   match[5],
			currency: match[6],
			client:   match[7],
			location: match[8],
		}
		wallet.itemizedTransactions = append(wallet.itemizedTransactions, item)
	}

	sort.Slice(wallet.transactions, func(i, j int) bool {
		return fmt.Sprintf("%v", wallet.transactions[i]) < fmt.Sprintf("%v", wallet.transactions[j])
	})
	sort.Slice(wallet.itemizedTransactions, func(i, j int) bool {
		return fmt.Sprintf("%v", wallet.itemizedTransactions[i]) < fmt.Sprintf("%v", wallet.itemizedTransactions[j])
	})

	return wallet, rest
}
