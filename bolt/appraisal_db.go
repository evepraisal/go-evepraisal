package bolt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
	"github.com/martinlindhe/base36"
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

		encodedID := EncodeAppraisalIDFromUint64(id)
		appraisal.ID, err = DecodeAppraisalID(encodedID)
		if err != nil {
			return err
		}

		appraisalBytes, err := json.Marshal(appraisal)
		if err != nil {
			return err
		}

		return b.Put(encodedID, appraisalBytes)
	})
}

func (db *AppraisalDB) GetAppraisal(appraisalID string) (*evepraisal.Appraisal, error) {
	dbID, err := EncodeAppraisalID(appraisalID)
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
	log.Println(reqCount, kind)
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
			id, _ := DecodeAppraisalID(k)
			log.Println(id, appraisal.CreatedTime())
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
		total = int64(b.Stats().KeyN)
		return nil
	})

	return total, err
}

func (db *AppraisalDB) Close() error {
	return db.db.Close()
}

func EncodeAppraisalID(appraisalID string) ([]byte, error) {
	// TODO: check for [a-z0-9] charset
	return EncodeAppraisalIDFromUint64(base36.Decode(appraisalID)), nil
}

func EncodeAppraisalIDFromUint64(appraisalID uint64) []byte {
	dbID := make([]byte, 8)
	binary.BigEndian.PutUint64(dbID, appraisalID)
	return dbID
}

func DecodeAppraisalID(dbID []byte) (string, error) {
	return strings.ToLower(base36.Encode(binary.BigEndian.Uint64(dbID))), nil
}
