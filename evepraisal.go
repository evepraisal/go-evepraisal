package evepraisal

import (
	"errors"
	"net/http"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/typedb"
	newrelic "github.com/newrelic/go-agent"
)

// App holds references to all of the app state that's needed. This is typically created in the 'evepraisal' package.
type App struct {
	AppraisalDB         AppraisalDB
	TypeDB              typedb.TypeDB
	PriceDB             PriceDB
	Parser              parsers.Parser
	WebContext          WebContext
	NewRelicApplication newrelic.Application
}

// AppraisalDB allows for creating, deleting and retreiving appraisals
type AppraisalDB interface {
	PutNewAppraisal(appraisal *Appraisal) error
	GetAppraisal(appraisalID string, updateUsedTime bool) (*Appraisal, error)
	LatestAppraisals(count int, kind string) ([]Appraisal, error)
	LatestAppraisalsByUser(user User, count int, kind string, after string) ([]Appraisal, error)
	TotalAppraisals() (int64, error)
	IncrementTotalAppraisals() error
	DeleteAppraisal(appraisalID string) error
	Close() error
}

var (
	// ErrAppraisalNotFound is returned whenever an appraisal by a given ID can't be found
	ErrAppraisalNotFound = errors.New("Appraisal not found")
)

// PriceDB holds prices for eve online items. Something else should update them
type PriceDB interface {
	GetPrice(market string, typeID int64) (Prices, bool)
	UpdatePrices([]MarketItemPrices) error
	Close() error
}

// TransactionLogger is used to log general events and HTTP requests
type TransactionLogger interface {
	StartTransaction(identifier string) Transaction
	StartWebTransaction(identifier string, w http.ResponseWriter, r *http.Request) Transaction
}

// Transaction is used to signal the normal or abnormal end to a given event
type Transaction interface {
	NoticeError(error) error
	End() error
}

// WebContext holds HTTP handlers and stuff
type WebContext interface {
	HTTPHandler() http.Handler
	Reload() error
}
