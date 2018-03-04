package bolt

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

var (
	// ErrCacheKeyNotFound is returned whenever there's a cache miss
	ErrCacheKeyNotFound = errors.New("Cache key not found, cache miss")
)

// HTTPCache caches HTTP responses and adheres to the interface in github.com/gregjones/httpcache
type HTTPCache struct {
	db *bolt.DB
}

// NewHTTPCache makes a new boltdb-based HTTPCache object
func NewHTTPCache(filename string) (*HTTPCache, error) {
	var (
		db  *bolt.DB
		err error
	)

	db, err = bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("httpcache"))
		if err != nil {
			return fmt.Errorf("create httpcache bucket: %s", err)
		}
		return nil
	})

	return &HTTPCache{db}, err
}

// Get retreives a cache entry, if it exists
func (c *HTTPCache) Get(key string) (resp []byte, ok bool) {
	var result []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("httpcache"))
		tmp := b.Get([]byte(key))

		if tmp == nil {
			return ErrCacheKeyNotFound
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

// Set creates a new cache entry and logs on failure
func (c *HTTPCache) Set(key string, resp []byte) {
	err := c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("httpcache")).Put([]byte(key), resp)
	})
	if err != nil {
		log.Println("Error: saving in httpcache: ", err)
	}
}

// Delete deletes a cache entry and logs on failure
func (c *HTTPCache) Delete(key string) {
	err := c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("httpcache")).Delete([]byte(key))
	})
	if err != nil {
		log.Println("Error: deleting in httpcache: ", err)
	}
}

// Close cleans up any open boltdb connections
func (c *HTTPCache) Close() error {
	return c.db.Close()
}
