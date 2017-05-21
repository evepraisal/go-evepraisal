package bolt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
	"github.com/golang/snappy"
)

type AppraisalDB struct {
	db   *bolt.DB
	wg   *sync.WaitGroup
	stop chan (bool)
}

func NewAppraisalDB(filename string) (evepraisal.AppraisalDB, error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("appraisals"))
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
		if err != bolt.ErrBucketExists {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	appraisalDB := &AppraisalDB{
		db:   db,
		wg:   &sync.WaitGroup{},
		stop: make(chan bool),
	}

	appraisalDB.wg.Add(1)
	go appraisalDB.startReaper()
	return appraisalDB, nil
}

func (db *AppraisalDB) PutNewAppraisal(appraisal *evepraisal.Appraisal) error {
	var dbID []byte
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		var err error
		if appraisal.ID == "" {
			id, err := b.NextSequence()
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

		appraisalBytes, err := json.Marshal(appraisal)
		if err != nil {
			return err
		}

		return b.Put(dbID, snappy.Encode(nil, appraisalBytes))
	})
	if err != nil {
		go db.setLastUsedTime(dbID)
	}
	return err
}

func (db *AppraisalDB) GetAppraisal(appraisalID string) (*evepraisal.Appraisal, error) {
	dbID, err := EncodeDBID(appraisalID)
	if err != nil {
		return nil, err
	}

	appraisal := &evepraisal.Appraisal{}

	err = db.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte("appraisals"))
		buf := b.Get(dbID)
		if buf == nil {
			return evepraisal.AppraisalNotFound
		}

		buf, err = snappy.Decode(nil, buf)
		if err != nil {
			return fmt.Errorf("Error when decoding: %s", err)
		}

		return json.Unmarshal(buf, appraisal)
	})

	go db.setLastUsedTime(dbID)

	return appraisal, err
}

func (db *AppraisalDB) LatestAppraisals(reqCount int, kind string) ([]evepraisal.Appraisal, error) {
	appraisals := make([]evepraisal.Appraisal, 0, reqCount)
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		c := b.Cursor()
		for key, val := c.Last(); key != nil; key, val = c.Prev() {
			appraisal := evepraisal.Appraisal{}

			buf, err := snappy.Decode(nil, val)
			if err != nil {
				return fmt.Errorf("Error when decoding: %s", err)
			}

			err = json.Unmarshal(buf, &appraisal)
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
		}

		return nil
	})

	return appraisals, err
}

func (db *AppraisalDB) TotalAppraisals() (int64, error) {
	var total int64
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		total = int64(b.Sequence())
		return nil
	})

	return total, err
}

func (db *AppraisalDB) Close() error {
	close(db.stop)
	db.wg.Wait()
	return db.db.Close()
}

func EncodeDBID(appraisalID string) ([]byte, error) {
	return EncodeDBIDFromUint64(evepraisal.AppraisalIDToUint64(appraisalID)), nil
}

func EncodeDBIDFromUint64(appraisalID uint64) []byte {
	dbID := make([]byte, 8)
	binary.BigEndian.PutUint64(dbID, appraisalID)
	return dbID
}

func DecodeDBID(dbID []byte) (string, error) {
	return strings.ToLower(evepraisal.Uint64ToAppraisalID(binary.BigEndian.Uint64(dbID))), nil
}

func (db *AppraisalDB) setLastUsedTime(dbID []byte) {
	now := time.Now().Unix()
	encodedNow := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedNow, uint64(now))
	err := db.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("appraisals-last-used")).Put(dbID, encodedNow)
	})

	if err != nil {
		log.Printf("WARNING: Error saving appraisal stats: %s", err)
	}
}

func (db *AppraisalDB) startReaper() {
	defer db.wg.Done()
	for {
		select {
		case <-db.stop:
			return
		default:
		}

		log.Println("Start reaping unused appraisals")
		unused := make([]string, 0)
		err := db.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("appraisals-last-used"))
			c := b.Cursor()
			for key, val := c.First(); key != nil; key, val = c.Next() {
				var timestamp time.Time
				if val != nil {
					timestamp = time.Unix(int64(binary.BigEndian.Uint64(val)), 0)
				} else {
					timestamp = time.Unix(0, 0)
				}

				log.Println(timestamp, time.Since(timestamp))
				if time.Since(timestamp) > time.Hour*24*90 {
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

		err = db.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("appraisals"))
			usedB := tx.Bucket([]byte("appraisals"))
			for _, appraisalID := range unused {
				dbID, err := EncodeDBID(appraisalID)
				if err != nil {
					log.Printf("Unable to parse appraisal ID (%s) %s", appraisalID, err)
					continue
				}

				err = b.Delete(dbID)
				if err != nil {
					return err
				}

				err = usedB.Delete(dbID)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("ERROR: Problem removing unused appraisals: %s", err)
		}

		log.Printf("Done reaping unused appraisals, removed %d appraisals", len(unused))
		time.Sleep(time.Hour)
	}
}
