package bolt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
)

type AppraisalDB struct {
	db *bolt.DB
}

func NewAppraisalDB(filename string) (evepraisal.AppraisalDB, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("appraisals"))
		if err != nil {
			return fmt.Errorf("create appraisal bucket: %s", err)
		}
		return nil
	})

	return &AppraisalDB{db: db}, err
}

func (db *AppraisalDB) PutNewAppraisal(appraisal *evepraisal.Appraisal) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("appraisals"))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		appraisal.ID = strconv.FormatUint(id, 10)
		appraisalBytes, err := json.Marshal(appraisal)
		if err != nil {
			return err
		}

		return b.Put(MakeDatabaseIDFromUint64(id), appraisalBytes)
	})
}

func (db *AppraisalDB) GetAppraisal(appraisalID string) (*evepraisal.Appraisal, error) {
	dbID, err := MakeDatabaseID(appraisalID)
	if err != nil {
		return nil, err
	}

	appraisal := &evepraisal.Appraisal{}

	err = db.db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("appraisals"))
		buf := b.Get(dbID)
		if buf == nil {
			return evepraisal.AppraisalNotFound
		}
		return json.Unmarshal(buf, appraisal)
	})

	return appraisal, err
}

func (db *AppraisalDB) LatestAppraisals(reqCount int, kind string) ([]evepraisal.Appraisal, error) {
	appraisals := make([]evepraisal.Appraisal, 0, reqCount)
	err := db.db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("appraisals"))
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			appraisal := evepraisal.Appraisal{}
			err := json.Unmarshal(v, &appraisal)
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

func (db *AppraisalDB) Close() error {
	return db.db.Close()
}

func MakeDatabaseID(appraisalID string) ([]byte, error) {
	uintID, err := strconv.ParseUint(appraisalID, 10, 64)
	if err != nil {
		return nil, err
	}

	return MakeDatabaseIDFromUint64(uintID), nil
}

func MakeDatabaseIDFromUint64(appraisalID uint64) []byte {
	dbID := make([]byte, 8)
	binary.BigEndian.PutUint64(dbID, appraisalID)
	return dbID
}
