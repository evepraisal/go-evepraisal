package evepraisal

import (
	"errors"
	"net/http"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/newrelic/go-agent"
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

// ListAppraisalsOptions is used to pass options into AppraisalDB.ListAppraisals
type ListAppraisalsOptions struct {
	User             *User
	Limit            int
	Kind             string
	StartAppraisalID string
	EndAppraisalID   string
	SortDirection    string
}

// AppraisalDB allows for creating, deleting and retreiving appraisals
type AppraisalDB interface {
	PutNewAppraisal(appraisal *Appraisal) error
	GetAppraisal(appraisalID string) (*Appraisal, error)
	ListAppraisals(opts ListAppraisalsOptions) ([]Appraisal, error)
	TotalAppraisals() (int64, error)
	DeleteAppraisal(appraisalID string) error
	Backup(dir string) error
	Close() error
}

var (
	// ErrAppraisalNotFound is returned whenever an appraisal by a given ID can't be found
	ErrAppraisalNotFound = errors.New("Appraisal not found")
)

// PriceDB holds prices for eve online items. Something else should update them
type PriceDB interface {
	GetPrice(market string, typeID int64) (Prices, bool)
	UpdatePrice(market string, typeID int64, prices Prices) error
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
