package newrelic

import (
	"net/http"

	"github.com/evepraisal/go-evepraisal"
	"github.com/newrelic/go-agent"
)

type TransactionLogger struct {
	application newrelic.Application
}

func NewTransactionLogger(appName string, key string) (evepraisal.TransactionLogger, error) {
	config := newrelic.NewConfig(appName, key)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		return nil, err
	}
	return TransactionLogger{app}, nil
}

func (l TransactionLogger) StartTransaction(identifier string) evepraisal.Transaction {
	return l.application.StartTransaction(identifier, nil, nil)
}

func (l TransactionLogger) StartWebTransaction(identifier string, w http.ResponseWriter, r *http.Request) evepraisal.Transaction {
	return l.application.StartTransaction(identifier, w, r)
}
