package bolt

import (
	"encoding/binary"
	"encoding/json"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/golang/snappy"
)

type TypeDB struct {
	db *bolt.DB
}

func NewTypeDB(filename string, writable bool) (typedb.TypeDB, error) {
	opts := &bolt.Options{Timeout: 1 * time.Second}
	var (
		db  *bolt.DB
		err error
	)
	if !writable {
		opts.ReadOnly = true
		db, err = bolt.Open(filename, 0600, opts)
		if err != nil {
			return nil, err
		}
	} else {
		db, err = bolt.Open(filename, 0600, opts)
		if err != nil {
			return nil, err
		}

		// Init our buckets in case this is a fresh DB
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("types_by_name"))
			if err != nil && err != bolt.ErrBucketExists {
				return err
			}

			_, err = tx.CreateBucket([]byte("types_by_id"))
			if err != nil && err != bolt.ErrBucketExists {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return &TypeDB{db: db}, err
}

func (db *TypeDB) GetType(typeName string) (typedb.EveType, bool) {
	evetype := typedb.EveType{}
	var buf []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		buf = tx.Bucket([]byte("types_by_name")).Get([]byte(strings.ToLower(typeName)))
		return nil
	})

	if buf == nil || err != nil {
		return evetype, false
	}

	buf, err = snappy.Decode(nil, buf)
	if err != nil {
		return evetype, false
	}

	err = json.Unmarshal(buf, &evetype)
	if err != nil {
		return evetype, false
	}

	return evetype, true
}

func (db *TypeDB) HasType(typeName string) bool {
	var buf []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		buf = tx.Bucket([]byte("types_by_name")).Get([]byte(strings.ToLower(typeName)))
		return nil
	})

	if buf == nil || err != nil {
		return false
	}

	return true
}

func (db *TypeDB) GetTypeByID(typeID int64) (typedb.EveType, bool) {
	encodedEveTypeID := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedEveTypeID, uint64(typeID))

	evetype := typedb.EveType{}
	var buf []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		buf = tx.Bucket([]byte("types_by_id")).Get(encodedEveTypeID)
		return nil
	})

	if buf == nil || err != nil {
		return evetype, false
	}

	buf, err = snappy.Decode(nil, buf)
	if err != nil {
		return evetype, false
	}

	err = json.Unmarshal(buf, &evetype)
	if err != nil {
		return evetype, false
	}

	return evetype, true
}

func (db *TypeDB) PutType(eveType typedb.EveType) error {
	typeBytes, err := json.Marshal(eveType)
	if err != nil {
		return err
	}
	typeBytes = snappy.Encode(nil, typeBytes)
	encodedEveTypeID := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedEveTypeID, uint64(eveType.ID))

	return db.db.Update(func(tx *bolt.Tx) error {
		byName := tx.Bucket([]byte("types_by_name"))
		err := byName.Put([]byte(strings.ToLower(eveType.Name)), typeBytes)
		if err != nil {
			return err
		}

		byID := tx.Bucket([]byte("types_by_id"))
		err = byID.Put(encodedEveTypeID, typeBytes)
		if err != nil {
			return err
		}

		return nil
	})
}

func (db *TypeDB) Close() error {
	return db.db.Close()
}
