package bolt

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	path string

	lock    *sync.RWMutex
	wg      *sync.WaitGroup
	stop    chan (bool)
	stopped bool
}

func openDB(filename string) (*bolt.DB, error) {
	var nmapSize = 0

	// Give 2GB of buffer space for the nmap (for backups)
	dbStat, err := os.Stat(filename)
	if err == nil {
		nmapSize = int(dbStat.Size()) + 2000000000
	}

	return bolt.Open(filename, 0600, &bolt.Options{
		Timeout:         1 * time.Second,
		InitialMmapSize: nmapSize,
	})
}

// NewAppraisalDB returns a new AppraisalDB with the buckets created
func NewAppraisalDB(filename string) (evepraisal.AppraisalDB, error) {
	db, err := openDB(filename)
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
		path: filename,
		lock: &sync.RWMutex{},
		wg:   &sync.WaitGroup{},
		stop: make(chan bool),
	}

	appraisalDB.wg.Add(1)
	go appraisalDB.startReaper()
	return appraisalDB, nil
}

// PutNewAppraisal stores the given appraisal
func (db *AppraisalDB) PutNewAppraisal(appraisal *evepraisal.Appraisal) error {
	db.lock.RLock()
	defer db.lock.RUnlock()

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
	db.lock.RLock()
	defer db.lock.RUnlock()

	var (
		dbID []byte
		err  error
		a    evepraisal.Appraisal
	)
	dbID, err = EncodeDBID(appraisalID)
	if err != nil {
		return nil, err
	}

	var appraisal *evepraisal.Appraisal
	err = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		buf := b.Get(dbID)
		if buf == nil {
			return evepraisal.ErrAppraisalNotFound
		}
		a, err = decodeAppraisal(buf)
		if err != nil {
			return err
		}
		appraisal = &a
		return nil
	})

	return appraisal, err
}

// ListAppraisals returns the latest appraisals by the given user
func (db *AppraisalDB) ListAppraisals(opts evepraisal.ListAppraisalsOptions) ([]evepraisal.Appraisal, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var (
		startingKey     string
		endingCondition func(string) bool
		seekTo          []byte
		nextFn          func(*bolt.Cursor) ([]byte, []byte)
	)

	if opts.SortDirection == "" || opts.SortDirection == "ASC" {
		startingKey = opts.StartAppraisalID
		endingCondition = func(key string) bool { return opts.EndAppraisalID != "" && key > opts.EndAppraisalID }
		nextFn = func(c *bolt.Cursor) ([]byte, []byte) { return c.Next() }
	} else if opts.SortDirection == "DESC" {
		startingKey = opts.EndAppraisalID
		endingCondition = func(key string) bool { return opts.StartAppraisalID != "" && key < opts.StartAppraisalID }
		nextFn = func(c *bolt.Cursor) ([]byte, []byte) { return c.Prev() }
	} else {
		return nil, fmt.Errorf("SortDirection is not valid: %s", opts.SortDirection)
	}

	if startingKey != "" {
		afterDBID, err := EncodeDBID(startingKey)
		if err != nil {
			return nil, err
		}
		seekTo = afterDBID
	} else {
		seekTo = []byte(";")
	}

	if opts.User != nil {
		// the appraisals-by-user bucket has keys formatted like "{character owner hash}:{encoded appraisal id}"
		b := bytes.NewBuffer([]byte(opts.User.CharacterOwnerHash))
		_ = b.WriteByte(':')   /* #nosec */
		_, _ = b.Write(seekTo) /* #nosec */
		seekTo = b.Bytes()
	}

	appraisals := make([]evepraisal.Appraisal, 0, opts.Limit)
	queriedCount := 0

	err := db.DB.View(func(tx *bolt.Tx) error {
		byUserBucket := tx.Bucket([]byte("appraisals-by-user"))
		byIDBucket := tx.Bucket([]byte("appraisals"))
		var (
			c              *bolt.Cursor
			keyConstraints func(string) bool
		)
		if opts.User == nil {
			c = byIDBucket.Cursor()
			keyConstraints = func(key string) bool { return key != "" }
		} else {
			c = byUserBucket.Cursor()
			keyConstraints = func(key string) bool { return strings.HasPrefix(string(key), opts.User.CharacterOwnerHash) }
		}

		c.Seek(seekTo)

		for key, val := nextFn(c); keyConstraints(string(key)); key, val = nextFn(c) {
			if opts.User != nil {
				val = byIDBucket.Get(val)
			}

			appraisal, err := decodeAppraisal(val)
			if err != nil {
				return err
			}

			if opts.Kind != "" && appraisal.Kind != opts.Kind {
				continue
			}

			if endingCondition(appraisal.ID) {
				break
			}

			appraisals = append(appraisals, appraisal)

			if opts.Limit > 0 {
				if len(appraisals) >= opts.Limit {
					break
				}

				if queriedCount >= opts.Limit*10 {
					break
				}
			}

			// Extra protection
			if opts.User != nil {
				if appraisal.User == nil {
					break
				}

				if appraisal.User.CharacterOwnerHash != opts.User.CharacterOwnerHash {
					break
				}
			}
		}

		return nil
	})

	return appraisals, err
}

// TotalAppraisals returns the number of total appraisals
func (db *AppraisalDB) TotalAppraisals() (int64, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

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
	db.lock.RLock()
	defer db.lock.RUnlock()

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

// resetDB is used to re-initialize the database because boltdb is weird and can't grow the database with
// pending readers.
func (db *AppraisalDB) resetDB() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	err := db.DB.Close()
	if err != nil {
		return err
	}

	db.DB, err = openDB(db.path)
	if err != nil {
		return err
	}

	return nil
}

// Backup writes out a backup to the given directory
func (db *AppraisalDB) Backup(dir string) error {
	log.Println("BACKUP: Backup to directory:", dir)

	// Load the meta file that tells us the latest appraisal ID so we know where to start
	var startAppraisalID string
	startAppraisalIDBytes, err := ioutil.ReadFile(filepath.Join(dir, "_meta"))
	if os.IsNotExist(err) {
		startAppraisalID = "0"
	} else if err != nil {
		return err
	} else {
		startAppraisalID = string(startAppraisalIDBytes)
	}

	// We might need a little time to query, if it takes too long, the db freezes up when trying to grow the db file
	err = db.resetDB()
	if err != nil {
		return err
	}

	fetchAndBackupAppraisals := func(appraisalID string) (string, error) {
		log.Println("BACKUP: starting new query at", appraisalID)
		limit := 10000
		opts := evepraisal.ListAppraisalsOptions{
			StartAppraisalID: appraisalID,
			Limit:            limit + 1,
			SortDirection:    "ASC",
		}

		var appraisals []evepraisal.Appraisal
		appraisals, err = db.ListAppraisals(opts)
		if err != nil {
			return "", err
		}

		next := ""
		if len(appraisals) > limit {
			next = appraisals[len(appraisals)-1].ID
			appraisals = appraisals[0:limit]
		}

		if len(appraisals) == 0 {
			log.Println("BACKUP: finished", appraisalID)
			return "", nil
		}

		first := appraisals[0].ID
		last := appraisals[len(appraisals)-1].ID
		filename := fmt.Sprintf("appraisals-%s-%s.json", first, last)
		filepath := filepath.Join(dir, filename)

		log.Printf("BACKUP: got %d appraisals for file %s", len(appraisals), filepath)
		var f *os.File
		f, err = os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", err
		}
		defer f.Close()
		cf := gzip.NewWriter(f)
		encoder := json.NewEncoder(cf)
		err = encoder.Encode(appraisals)
		if err != nil {
			return "", err
		}

		log.Printf("BACKUP: created backed up appraisals: %s", filepath)

		return next, nil
	}

	next := startAppraisalID
	for next != "" {
		next, err = fetchAndBackupAppraisals(next)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database
func (db *AppraisalDB) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.stopped = true
	close(db.stop)
	db.wg.Wait()
	return db.DB.Close()
}

func decodeAppraisal(val []byte) (evepraisal.Appraisal, error) {
	appraisal := evepraisal.Appraisal{}

	buf, err := snappy.Decode(nil, val)
	if err != nil {
		return appraisal, fmt.Errorf("Error when decoding: %s", err)
	}

	decoder := gob.NewDecoder(bytes.NewBuffer(buf))
	if err != nil {
		return appraisal, fmt.Errorf("Error when decoding: %s", err)
	}

	err = decoder.Decode(&appraisal)
	return appraisal, err

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
