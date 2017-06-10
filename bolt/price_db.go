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

type PriceDB struct {
	db *bolt.DB
}

func NewPriceDB(filename string) (evepraisal.PriceDB, error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("prices"))
		if err != nil {
			return fmt.Errorf("create prices bucket: %s", err)
		}
		return nil
	})

	return &PriceDB{db: db}, err
}

func (db *PriceDB) GetPrice(market string, typeID int64) (evepraisal.Prices, bool) {
	prices := &evepraisal.Prices{}

	err := db.db.View(func(tx *bolt.Tx) error {
		var err error
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

func (db *PriceDB) UpdatePrice(market string, typeID int64, prices evepraisal.Prices) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("prices"))
		priceBytes, err := json.Marshal(prices)
		if err != nil {
			return err
		}

		return b.Put([]byte(fmt.Sprintf("%s|%d", market, typeID)), snappy.Encode(nil, priceBytes))
	})
}

func (db *PriceDB) Close() error {
	return db.db.Close()
}
