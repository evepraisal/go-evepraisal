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

var expireCheckDuration = time.Hour * 24 * 80
var maxExpireTime = time.Hour * 24 * 90

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
		var totalAppraisals uint64
		b, err = tx.CreateBucket([]byte("appraisals"))
		if err == nil {
			err = b.SetSequence(0)
			if err != nil {
				return fmt.Errorf("set appraisal bucket sequence: %s", err)
			}
			log.Println("Appraisal bucket created")
		} else if err != bolt.ErrBucketExists {
			return err
		}
		totalAppraisals = tx.Bucket([]byte("appraisals")).Sequence()

		_, err = tx.CreateBucket([]byte("appraisals-last-used"))
		if err != nil && err != bolt.ErrBucketExists {
			return err
		}

		_, err = tx.CreateBucket([]byte("appraisals-by-user"))
		if err != nil && err != bolt.ErrBucketExists {
			return err
		}

		b, err = tx.CreateBucket([]byte("stats"))
		if err == nil {
			err = putTotalAppraisals(tx, totalAppraisals)
			log.Printf("Stats bucket created at total_appraisals=%d", totalAppraisals)
			return err
		} else if err != bolt.ErrBucketExists {
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
	if err == nil {
		go db.setLastUsedTime(dbID)
		go db.IncrementTotalAppraisals()
	}
	return err
}

// GetAppraisal returns the given appraisal by ID
func (db *AppraisalDB) GetAppraisal(appraisalID string, updateUsedTime bool) (*evepraisal.Appraisal, error) {
	appraisal, err := db.getAppraisal(appraisalID)
	if err != nil {
		return nil, err
	}

	dbID, err := EncodeDBID(appraisalID)
	if err != nil {
		return nil, err
	}
	if updateUsedTime {
		go db.setLastUsedTime(dbID)
	}

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

func readTotalAppraisals(tx *bolt.Tx) uint64 {
	b := tx.Bucket([]byte("stats"))
	return binary.BigEndian.Uint64(b.Get([]byte("total_appraisals")))
}

func putTotalAppraisals(tx *bolt.Tx, value uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, value)
	return tx.Bucket([]byte("stats")).Put([]byte("total_appraisals"), buf)
}

// TotalAppraisals returns the number of total appraisals
func (db *AppraisalDB) TotalAppraisals() (int64, error) {
	var total int64
	err := db.DB.View(func(tx *bolt.Tx) error {
		total = int64(readTotalAppraisals(tx))
		return nil
	})

	return total, err
}

// IncrementTotalAppraisals will increment to total appraisal count, useful to track non-persistent appraisals
func (db *AppraisalDB) IncrementTotalAppraisals() error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		total := readTotalAppraisals(tx)
		return putTotalAppraisals(tx, total+1)
	})

	return err
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
		log.Println("Start reaping deletable appraisals")
		var (
			deletable      = make([]string, 0)
			appraisalCount = 0
			now            = time.Now()
			appraisal      *evepraisal.Appraisal
			sleepTime      = time.Hour
		)
		err := db.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("appraisals-last-used"))
			c := b.Cursor()
			for key, val := c.First(); key != nil; key, val = c.Next() {
				appraisalCount++
				if appraisalCount%1000 == 0 {
					select {
					case <-db.stop:
						return nil
					default:
					}
				}

				var usedTimestamp time.Time
				if val != nil {
					usedTimestamp = time.Unix(int64(binary.BigEndian.Uint64(val)), 0)
				} else {
					usedTimestamp = time.Unix(0, 0)
				}

				// Only look closely at appraisals that haven't been used in expireCheckDuration
				if now.Sub(usedTimestamp) > expireCheckDuration {
					appraisalID, err := DecodeDBID(key)
					if err != nil {
						log.Printf("Unable to parse appraisal ID (%s) %s", appraisalID, err)
						continue
					}

					if now.Sub(usedTimestamp) > maxExpireTime {
						deletable = append(deletable, appraisalID)
						continue
					}

					appraisal, err = db.getAppraisal(appraisalID)
					if err != nil {
						log.Printf("Unable to parse appraisal (%s) %s", appraisalID, err)
						continue
					}

					if appraisal.IsExpired(now, usedTimestamp) {
						deletable = append(deletable, appraisalID)
					}

					if len(deletable) > 1000 {
						log.Println("Too many to delete in one go. Breaking early and re-trigger another reaper")
						sleepTime = 5 * time.Second
						break
					}
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("ERROR: Problem querying for deletable appraisals: %s", err)
		}

		log.Printf("Reaper starting to delete %d appraisals", len(deletable))

		for i, appraisalID := range deletable {
			if i%100 == 0 {
				select {
				case <-db.stop:
					return
				default:
				}
			}

			err = db.DeleteAppraisal(appraisalID)
			if err != nil {
				log.Printf("ERROR: Problem removing deletable appraisals: %s", err)
			}
		}

		log.Printf("Done reaping deletable appraisals, removed %d (out of %d) appraisals. Sleeping for %s", len(deletable), appraisalCount, sleepTime)
		select {
		case <-db.stop:
			return
		case <-time.After(sleepTime):
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
