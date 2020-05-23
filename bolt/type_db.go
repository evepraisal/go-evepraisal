package bolt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	// Imported to register the standard analysis
	_ "github.com/blevesearch/bleve/analysis/analyzer/standard"
	// Imported to register boltdb with bleve
	_ "github.com/blevesearch/bleve/index/store/boltdb"
	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/golang/snappy"
)

// TypeDB holds all EveTypes
type TypeDB struct {
	db            *bolt.DB
	index         bleve.Index
	filename      string
	indexFilename string
}

var aliases = map[string]string{
	"skill injector": "large skill injector",
}

// NewTypeDB returns a new TypeDB
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
			_, err = tx.CreateBucket([]byte("types_by_name"))
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
	if _, err = os.Stat(indexFilename); os.IsNotExist(err) {
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
			if err != nil {
				return nil, err
			}
		}

	} else {
		return nil, err
	}

	return &TypeDB{db: db, index: index, filename: filename, indexFilename: indexFilename}, err
}

func massageTypeName(typeName string) string {
	typeName = strings.ToLower(typeName)
	aliasedTypeName, ok := aliases[typeName]
	if ok {
		typeName = aliasedTypeName
	}

	if strings.HasSuffix(typeName, "'s frozen corpse") {
		return "frozen corpse"
	}
	return typeName
}

// GetType returns the EveType given a name
func (db *TypeDB) GetType(typeName string) (typedb.EveType, bool) {
	evetype := typedb.EveType{}
	var buf []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		buf = tx.Bucket([]byte("types_by_name")).Get([]byte(massageTypeName(typeName)))
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

// HasType returns whether or not the type exists given a name
func (db *TypeDB) HasType(typeName string) bool {
	var buf []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		buf = tx.Bucket([]byte("types_by_name")).Get([]byte(massageTypeName(typeName)))
		return nil
	})

	if buf == nil || err != nil {
		return false
	}

	return true
}

// GetTypeByID returns the EveType that matches the integer ID
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

// ListTypes returns all the types
func (db *TypeDB) ListTypes(startingTypeID int64, limit int64) ([]typedb.EveType, error) {
	encodedStartingTypeID := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedStartingTypeID, uint64(startingTypeID))

	items := make([]typedb.EveType, 0)
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("types_by_id"))
		c := b.Cursor()
		c.Seek(encodedStartingTypeID)
		var (
			buf []byte
			err error
		)
		for key, val := c.Next(); key != nil; key, val = c.Next() {
			evetype := typedb.EveType{}
			buf, err = snappy.Decode(nil, val)
			if err != nil {
				return err
			}

			err = json.Unmarshal(buf, &evetype)
			if err != nil {
				return err
			}
			items = append(items, evetype)

			if int64(len(items)) >= limit {
				return nil
			}
		}
		return nil
	})

	return items, err
}

// PutTypes will insert/update the given EveTypes
func (db *TypeDB) PutTypes(eveTypes []typedb.EveType) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		for _, eveType := range eveTypes {

			typeBytes, err := json.Marshal(eveType)
			if err != nil {
				return err
			}
			typeBytes = snappy.Encode(nil, typeBytes)
			encodedEveTypeID := make([]byte, 8)
			binary.BigEndian.PutUint64(encodedEveTypeID, uint64(eveType.ID))

			// NOTE - only index off-market items by name if it's not going to override another type
			byName := tx.Bucket([]byte("types_by_name"))
			skipByName := eveType.MarketGroupID == 0 && db.HasType(eveType.Name)
			if !skipByName {
				err = byName.Put([]byte(strings.ToLower(eveType.Name)), typeBytes)
				if err != nil {
					return err
				}
			}
			for _, alias := range eveType.Aliases {
				skipByName := eveType.MarketGroupID == 0 && db.HasType(alias)
				if !skipByName {
					err = byName.Put([]byte(strings.ToLower(alias)), typeBytes)
					if err != nil {
						return err
					}
				}
			}

			byID := tx.Bucket([]byte("types_by_id"))
			err = byID.Put(encodedEveTypeID, typeBytes)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	batch := db.index.NewBatch()
	for _, eveType := range eveTypes {
		err := batch.Index(strconv.FormatInt(eveType.ID, 10), eveType.Name)
		if err != nil {
			return err
		}
	}

	return db.index.Batch(batch)
}

// Search allows for searching based on an incomplete name of a type
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

	searchRequest := bleve.NewSearchRequestOptions(q, 20, 0, false)
	searchResults, err := db.index.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	results := make([]typedb.EveType, len(searchResults.Hits))
	for i, result := range searchResults.Hits {
		id, err := strconv.ParseInt(result.ID, 10, 64)
		if err != nil {
			log.Println("error parsing the search ID into an integer", err)
		}
		t, _ := db.GetTypeByID(id)
		results[i] = t
	}

	return results
}

// Delete will delete the entire type DB
func (db *TypeDB) Delete() error {
	err := os.RemoveAll(db.filename)
	if err != nil {
		return err
	}

	return os.RemoveAll(db.indexFilename)
}

// Close will close the type database
func (db *TypeDB) Close() error {
	err := db.db.Close()
	if err != nil {
		return err
	}
	return db.index.Close()
}
