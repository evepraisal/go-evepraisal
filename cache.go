package evepraisal

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

func SaveToCache(db *leveldb.DB, key string, val interface{}) error {
	v, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put([]byte(key), v, nil)
}

func GetFromCache(db *leveldb.DB, key string, val interface{}) error {
	v, err := db.Get([]byte(key), nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(v, val)
}
