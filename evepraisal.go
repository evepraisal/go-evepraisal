package evepraisal

import (
	"errors"

	"github.com/evepraisal/go-evepraisal/parsers"
)

type App struct {
	AppraisalDB AppraisalDB
	CacheDB     CacheDB
	TypeDB      TypeDB
	PriceDB     PriceDB
	Parser      parsers.Parser
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
	Close() error
}

var (
	AppraisalNotFound = errors.New("Appraisal not found")
)

type TypeDB interface {
	GetType(typeName string) (EveType, bool)
	HasType(typeName string) bool
	Close() error
}

type PriceDB interface {
	GetPrice(market string, typeID int64) (Prices, bool)
	Close() error
}

type EveType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
