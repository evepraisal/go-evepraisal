package bolt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	// Imported to register boltdb with bleve
	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/analyzer/standard"
	_ "github.com/blevesearch/bleve/index/store/boltdb"
	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/golang/snappy"
)

type TypeDB struct {
	db            *bolt.DB
	index         bleve.Index
	filename      string
	indexFilename string
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

	var index bleve.Index
	indexFilename := filename + ".index"
	if _, err := os.Stat(indexFilename); os.IsNotExist(err) {
		if !writable {
			return nil, fmt.Errorf("Index (%s) does not exist so it cannot be opened in read-only mode", indexFilename)
		}
		mapping := bleve.NewIndexMapping()
		mapping.DefaultAnalyzer = "standard"
		index, err = bleve.New(indexFilename, mapping)
		if err != nil {
			return nil, err
		}
	} else if err == nil {
		if writable {
			index, err = bleve.Open(indexFilename)
			if err != nil {
				return nil, err
			}
		} else {
			index, err = bleve.OpenUsing(indexFilename, map[string]interface{}{
				"read_only": true,
			})
		}

	} else {
		return nil, err
	}

	return &TypeDB{db: db, index: index, filename: filename, indexFilename: indexFilename}, err
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

	err = db.db.Update(func(tx *bolt.Tx) error {
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
	if err != nil {
		return err
	}

	return db.index.Index(strconv.FormatInt(eveType.ID, 10), eveType.Name)
}

func (db *TypeDB) Search(s string) []typedb.EveType {
	searchString := strings.ToLower(s)

	// First try an exact match
	t, ok := db.GetType(searchString)
	if ok {
		return []typedb.EveType{t}
	}

	// Then try a real search
	q1 := bleve.NewTermQuery(searchString)
	q1.SetBoost(10)

	q2 := bleve.NewPrefixQuery(searchString)
	q2.SetBoost(5)

	q3 := bleve.NewMatchPhraseQuery(searchString)

	q := bleve.NewDisjunctionQuery(q1, q2, q3)

	searchRequest := bleve.NewSearchRequest(q)
	searchResults, err := db.index.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	results := make([]typedb.EveType, len(searchResults.Hits))
	for i, result := range searchResults.Hits {
		id, _ := strconv.ParseInt(result.ID, 10, 64)
		t, _ := db.GetTypeByID(id)
		results[i] = t
	}

	return results
}

func (db *TypeDB) Delete() error {
	err := os.RemoveAll(db.filename)
	if err != nil {
		return err
	}

	return os.RemoveAll(db.indexFilename)
}

func (db *TypeDB) Close() error {
	err := db.db.Close()
	if err != nil {
		return err
	}
	return db.index.Close()
}
