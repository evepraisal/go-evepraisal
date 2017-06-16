package evepraisal

import (
	"errors"
	"net/http"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/newrelic/go-agent"
)

type App struct {
	AppraisalDB         AppraisalDB
	CacheDB             CacheDB
	TypeDB              typedb.TypeDB
	PriceDB             PriceDB
	Parser              parsers.Parser
	WebContext          WebContext
	NewRelicApplication newrelic.Application
}

type CacheDB interface {
	Put(key string, val []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Close() error
}

type AppraisalDB interface {
	PutNewAppraisal(appraisal *Appraisal) error
	GetAppraisal(appraisalID string) (*Appraisal, error)
	LatestAppraisals(count int, kind string) ([]Appraisal, error)
	LatestAppraisalsByUser(user User, count int, kind string) ([]Appraisal, error)
	TotalAppraisals() (int64, error)
	DeleteAppraisal(appraisalID string) error
	Close() error
}

var (
	AppraisalNotFound = errors.New("Appraisal not found")
)

type PriceDB interface {
	GetPrice(market string, typeID int64) (Prices, bool)
	UpdatePrice(market string, typeID int64, prices Prices) error
	Close() error
}

type TransactionLogger interface {
	StartTransaction(identifier string) Transaction
	StartWebTransaction(identifier string, w http.ResponseWriter, r *http.Request) Transaction
}

type Transaction interface {
	NoticeError(error) error
	End() error
}

type TrackingCodeProvider interface {
	TrackingJS() string
}

type WebContext interface {
	HTTPHandler() http.Handler
	Reload() error
}
