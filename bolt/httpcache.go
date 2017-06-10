package bolt

import (
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var (
	CacheKeyNotFound = errors.New("Cache key not found, cache miss")
)

type HTTPCache struct {
	db *bolt.DB
}

func NewHTTPCache(filename string) (*HTTPCache, error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("httpcache"))
		if err != nil {
			return fmt.Errorf("create httpcache bucket: %s", err)
		}
		return nil
	})

	return &HTTPCache{db}, err
}

func (c *HTTPCache) Get(key string) (resp []byte, ok bool) {
	var result []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("httpcache"))
		tmp := b.Get([]byte(key))

		if tmp == nil {
			return CacheKeyNotFound
		}

		// we need to copy the byte array because it might be re-used outside of this view function
		result = make([]byte, len(tmp))
		copy(result, tmp)
		return nil
	})
	if err != nil || result == nil {
		return nil, false
	}

	return result, true
}

func (c *HTTPCache) Set(key string, resp []byte) {
	c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("httpcache")).Put([]byte(key), resp)
	})
}

func (c *HTTPCache) Delete(key string) {
	c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("httpcache")).Delete([]byte(key))
	})
}

func (c *HTTPCache) Close() error {
	return c.db.Close()
}
