package noop

import (
	"net/http"

	"github.com/evepraisal/go-evepraisal"
)

type TransactionLogger struct{}

func NewTransactionLogger() evepraisal.TransactionLogger {
	return TransactionLogger{}
}

func (l TransactionLogger) StartTransaction(identifier string) evepraisal.Transaction {
	return Transaction{}
}

func (l TransactionLogger) StartWebTransaction(identifier string, w http.ResponseWriter, r *http.Request) evepraisal.Transaction {
	return Transaction{}
}

type Transaction struct{}

func (l Transaction) End() error { return nil }
