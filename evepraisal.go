package evepraisal

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/typedb"
)

type App struct {
	AppraisalDB       AppraisalDB
	CacheDB           CacheDB
	TypeDB            typedb.TypeDB
	PriceDB           PriceDB
	Parser            parsers.Parser
	TransactionLogger TransactionLogger
	ExtraJS           string
	templates         map[string]*template.Template
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
	TotalAppraisals() (int64, error)
	Close() error
}

var (
	AppraisalNotFound = errors.New("Appraisal not found")
)

type PriceDB interface {
	GetPrice(market string, typeID int64) (Prices, bool)
	Close() error
}

type TransactionLogger interface {
	StartTransaction(identifier string) Transaction
	StartWebTransaction(identifier string, w http.ResponseWriter, r *http.Request) Transaction
}

type Transaction interface {
	End() error
}

type TrackingCodeProvider interface {
	TrackingJS() string
}
