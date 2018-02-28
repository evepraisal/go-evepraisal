package bolt

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
	"github.com/golang/snappy"
)

var expireTime = time.Hour * 24 * 90

// AppraisalDB holds all appraisals
type AppraisalDB struct {
	DB   *bolt.DB
	wg   *sync.WaitGroup
	stop chan (bool)
}

// NewAppraisalDB returns a new AppraisalDB with the buckets created
func NewAppraisalDB(filename string) (evepraisal.AppraisalDB, error) {
	var nmapSize = 0

	// Give 2GB of buffer space for the nmap (for backups)
	dbStat, err := os.Stat(filename)
	if err == nil {
		nmapSize = int(dbStat.Size()) + 2000000000
	}

	db, err := bolt.Open(filename, 0600, &bolt.Options{
		Timeout:         1 * time.Second,
		InitialMmapSize: nmapSize,
	})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		b, err = tx.CreateBucket([]byte("appraisals"))
		if err == nil {
			err = b.SetSequence(20000000)
			if err != nil {
				return fmt.Errorf("set appraisal bucket sequence: %s", err)
			}
			log.Println("Appraisal bucket created")
		} else if err != bolt.ErrBucketExists {
			return err
		}

		_, err = tx.CreateBucket([]byte("appraisals-last-used"))
		if err != nil && err != bolt.ErrBucketExists {
			return err
		}

		_, err = tx.CreateBucket([]byte("appraisals-by-user"))
		if err != nil && err != bolt.ErrBucketExists {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	appraisalDB := &AppraisalDB{
		DB:   db,
		wg:   &sync.WaitGroup{},
		stop: make(chan bool),
	}

	appraisalDB.wg.Add(1)
	go appraisalDB.startReaper()
	return appraisalDB, nil
}

// PutNewAppraisal stores the given appraisal
func (db *AppraisalDB) PutNewAppraisal(appraisal *evepraisal.Appraisal) error {
	var dbID []byte
	err := db.DB.Update(func(tx *bolt.Tx) error {
		byIDBucket := tx.Bucket([]byte("appraisals"))
		var (
			err error
			id  uint64
		)

		if appraisal.ID == "" {
			id, err = byIDBucket.NextSequence()
			if err != nil {
				return err
			}

			dbID = EncodeDBIDFromUint64(id)
			appraisal.ID, err = DecodeDBID(dbID)
			if err != nil {
				return err
			}
		} else {
			dbID, err = EncodeDBID(appraisal.ID)
			if err != nil {
				return err
			}
		}
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err = encoder.Encode(appraisal)
		if err != nil {
			return err
		}

		err = byIDBucket.Put(dbID, snappy.Encode(nil, buf.Bytes()))
		if err != nil {
			return err
		}

		if appraisal.User != nil {
			byUserBucket := tx.Bucket([]byte("appraisals-by-user"))
			return byUserBucket.Put(append([]byte(fmt.Sprintf("%s:", appraisal.User.CharacterOwnerHash)), dbID...), dbID)
		}
		return nil
	})
	if err != nil {
		go db.setLastUsedTime(dbID)
	}
	return err
}

// GetAppraisal returns the given appraisal by ID
func (db *AppraisalDB) GetAppraisal(appraisalID string) (*evepraisal.Appraisal, error) {
	appraisal, err := db.getAppraisal(appraisalID)
	if err != nil {
		return nil, err
	}

	dbID, err := EncodeDBID(appraisalID)
	if err != nil {
		return nil, err
	}
	go db.setLastUsedTime(dbID)

	return appraisal, err
}

func (db *AppraisalDB) getAppraisal(appraisalID string) (*evepraisal.Appraisal, error) {
	var (
		dbID []byte
		err  error
	)
	dbID, err = EncodeDBID(appraisalID)
	if err != nil {
		return nil, err
	}

	appraisal := &evepraisal.Appraisal{}

	err = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		buf := b.Get(dbID)
		if buf == nil {
			return evepraisal.ErrAppraisalNotFound
		}

		buf, err = snappy.Decode(nil, buf)
		if err != nil {
			return fmt.Errorf("Error when decoding: %s", err)
		}

		decoder := gob.NewDecoder(bytes.NewBuffer(buf))
		return decoder.Decode(appraisal)
	})

	return appraisal, err
}

// LatestAppraisals returns the global latest appraisals
func (db *AppraisalDB) LatestAppraisals(reqCount int, kind string) ([]evepraisal.Appraisal, error) {
	appraisals := make([]evepraisal.Appraisal, 0, reqCount)
	queriedCount := 0
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		c := b.Cursor()
		for key, val := c.Last(); key != nil; key, val = c.Prev() {
			appraisal := evepraisal.Appraisal{}

			buf, err := snappy.Decode(nil, val)
			if err != nil {
				return fmt.Errorf("Error when decoding: %s", err)
			}

			decoder := gob.NewDecoder(bytes.NewBuffer(buf))
			err = decoder.Decode(&appraisal)
			if err != nil {
				return err
			}

			if appraisal.Private {
				continue
			}

			if kind != "" && appraisal.Kind != kind {
				continue
			}

			appraisals = append(appraisals, appraisal)

			if len(appraisals) >= reqCount {
				break
			}

			if queriedCount >= reqCount*10 {
				break
			}
		}

		return nil
	})

	return appraisals, err
}

// LatestAppraisalsByUser returns the latest appraisals by the given user
func (db *AppraisalDB) LatestAppraisalsByUser(user evepraisal.User, reqCount int, kind string, after string) ([]evepraisal.Appraisal, error) {
	appraisals := make([]evepraisal.Appraisal, 0, reqCount)
	queriedCount := 0
	err := db.DB.View(func(tx *bolt.Tx) error {
		byUserBucket := tx.Bucket([]byte("appraisals-by-user"))
		byIDBucket := tx.Bucket([]byte("appraisals"))
		c := byUserBucket.Cursor()

		var suffix []byte
		if after != "" {
			afterDBID, err := EncodeDBID(after)
			if err != nil {
				return err
			}
			suffix = append([]byte(":"), afterDBID...)
		} else {
			suffix = []byte(";")
		}

		c.Seek(append([]byte(user.CharacterOwnerHash), suffix...))

		for key, val := c.Prev(); strings.HasPrefix(string(key), user.CharacterOwnerHash); key, val = c.Prev() {
			buf, err := snappy.Decode(nil, byIDBucket.Get(val))
			if err != nil {
				return fmt.Errorf("Error when decoding: %s", err)
			}

			appraisal := evepraisal.Appraisal{}
			decoder := gob.NewDecoder(bytes.NewBuffer(buf))
			err = decoder.Decode(&appraisal)
			if err != nil {
				return err
			}

			if kind != "" && appraisal.Kind != kind {
				continue
			}

			appraisals = append(appraisals, appraisal)

			if len(appraisals) >= reqCount {
				break
			}

			if queriedCount >= reqCount*10 {
				break
			}
		}

		return nil
	})

	return appraisals, err
}

// TotalAppraisals returns the number of total appraisals
func (db *AppraisalDB) TotalAppraisals() (int64, error) {
	var total int64
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		total = int64(b.Sequence())
		return nil
	})

	return total, err
}

// DeleteAppraisal deletes an appraisal by ID
func (db *AppraisalDB) DeleteAppraisal(appraisalID string) error {
	appraisal, err := db.getAppraisal(appraisalID)
	appraisalFound := true
	if err == evepraisal.ErrAppraisalNotFound {
		appraisalFound = true
	} else if err != nil {
		return err
	}

	return db.DB.Update(func(tx *bolt.Tx) error {
		byIDBucket := tx.Bucket([]byte("appraisals"))
		byUserBucket := tx.Bucket([]byte("appraisals-by-user"))
		lastUsedB := tx.Bucket([]byte("appraisals-last-used"))
		dbID, err := EncodeDBID(appraisalID)
		if err != nil {
			return err
		}

		if appraisalFound && appraisal.User != nil {
			err = byUserBucket.Delete(append([]byte(fmt.Sprintf("%s:", appraisal.User.CharacterOwnerHash)), dbID...))
			if err != nil {
				return err
			}
		}

		err = byIDBucket.Delete(dbID)
		if err != nil {
			return err
		}

		err = lastUsedB.Delete(dbID)
		if err != nil {
			return err
		}
		return nil
	})
}

// Close closes the database
func (db *AppraisalDB) Close() error {
	close(db.stop)
	db.wg.Wait()
	return db.DB.Close()
}

func (db *AppraisalDB) setLastUsedTime(dbID []byte) {
	now := time.Now().Unix()
	encodedNow := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedNow, uint64(now))
	err := db.DB.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("appraisals-last-used")).Put(dbID, encodedNow)
	})

	if err != nil {
		log.Printf("WARNING: Error saving appraisal stats: %s", err)
	}
}

func (db *AppraisalDB) startReaper() {
	defer db.wg.Done()
	for {
		log.Println("Start reaping unused appraisals")
		unused := make([]string, 0)
		appraisalCount := 0
		err := db.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("appraisals-last-used"))
			c := b.Cursor()
			for key, val := c.First(); key != nil; key, val = c.Next() {
				appraisalCount++
				if appraisalCount%10000 == 0 {
					select {
					case <-db.stop:
						return nil
					}
				}

				var timestamp time.Time
				if val != nil {
					timestamp = time.Unix(int64(binary.BigEndian.Uint64(val)), 0)
				} else {
					timestamp = time.Unix(0, 0)
				}

				if time.Since(timestamp) > expireTime {
					appraisalID, err := DecodeDBID(key)
					if err != nil {
						log.Printf("Unable to parse appraisal ID (%s) %s", appraisalID, err)
						continue
					}
					unused = append(unused, appraisalID)
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("ERROR: Problem querying for unused appraisals: %s", err)
		}

		for _, appraisalID := range unused {
			err = db.DeleteAppraisal(appraisalID)
			if err != nil {
				log.Printf("ERROR: Problem removing unused appraisals: %s", err)
			}
		}

		log.Printf("Done reaping unused appraisals, removed %d (out of %d) appraisals", len(unused), appraisalCount)

		select {
		case <-db.stop:
			return
		case <-time.After(time.Hour):
		}
	}
}

// EncodeDBID encodes an appraisalID (which is seen by users) into a Unint64 that is used to sort appraisals properly
func EncodeDBID(appraisalID string) ([]byte, error) {
	return EncodeDBIDFromUint64(evepraisal.AppraisalIDToUint64(appraisalID)), nil
}

// EncodeDBIDFromUint64 converts the given uint64 into a byte array for storage. The uint64 is an intermediary form
// and is only really used when a new appraisalID is generated.
func EncodeDBIDFromUint64(appraisalID uint64) []byte {
	dbID := make([]byte, 8)
	binary.BigEndian.PutUint64(dbID, appraisalID)
	return dbID
}

// DecodeDBID converts the database ID into the user-visible appraisalID
func DecodeDBID(dbID []byte) (string, error) {
	return strings.ToLower(evepraisal.Uint64ToAppraisalID(binary.BigEndian.Uint64(dbID))), nil
}
