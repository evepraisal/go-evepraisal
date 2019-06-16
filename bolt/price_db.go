package bolt

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
	"github.com/golang/snappy"
)

// PriceDB stores the market prices for items
type PriceDB struct {
	db *bolt.DB
}

// NewPriceDB returns a new PriceDB instance
func NewPriceDB(filename string) (evepraisal.PriceDB, error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("prices"))
		if err != nil {
			return fmt.Errorf("create prices bucket: %s", err)
		}
		return nil
	})

	return &PriceDB{db: db}, err
}

// GetPrice returns the price for a type given a market name and typeID
func (db *PriceDB) GetPrice(market string, typeID int64) (evepraisal.Prices, bool) {
	prices := &evepraisal.Prices{}

	var err error
	err = db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("prices"))
		buf := b.Get([]byte(fmt.Sprintf("%s|%d", market, typeID)))
		if buf == nil {
			return errors.New("Price not found")
		}

		buf, err = snappy.Decode(nil, buf)
		if err != nil {
			return fmt.Errorf("Error when decoding: %s", err)
		}

		return json.Unmarshal(buf, prices)
	})

	if err != nil {
		return *prices, false
	}

	return *prices, true

}

// UpdatePrices updates the price for the given typeID in the given market
func (db *PriceDB) UpdatePrices(items []evepraisal.MarketItemPrices) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("prices"))

		for _, item := range items {
			priceBytes, err := json.Marshal(item.Prices)
			if err != nil {
				return err
			}

			err = b.Put([]byte(fmt.Sprintf("%s|%d", item.Market, item.TypeID)), snappy.Encode(nil, priceBytes))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Close cleans up the PriceDB
func (db *PriceDB) Close() error {
	return db.db.Close()
}
