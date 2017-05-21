package bolt

import (
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
)

var (
	CacheKeyNotFound = errors.New("Cache key not found, cache miss")
)

type CacheDB struct {
	db *bolt.DB
}

func NewCacheDB(filename string) (evepraisal.CacheDB, error) {

	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("cache"))
		if err != nil {
			return fmt.Errorf("create cache bucket: %s", err)
		}
		return nil
	})

	return &CacheDB{db: db}, err
}

func (c *CacheDB) Put(key string, value []byte) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("cache")).Put([]byte(key), value)
	})
}

func (c *CacheDB) Get(key string) ([]byte, error) {
	var result []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		tmp := b.Get([]byte(key))

		if tmp == nil {
			return CacheKeyNotFound
		}

		// we need to copy the byte array because it might be re-used outside of this view function
		result = make([]byte, len(tmp))
		copy(result, tmp)
		return nil
	})

	return result, err
}

func (c *CacheDB) Delete(key string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("cache")).Delete([]byte(key))
	})
}

func (c *CacheDB) Close() error {
	return c.db.Close()
}
