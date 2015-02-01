package boltorm

import (
	"encoding/binary"

	"github.com/boltdb/bolt"
)

type boltDB struct {
	db *bolt.DB
}

func NewBoltDB(db *bolt.DB) DB {
	return &boltDB{
		db: db,
	}
}

func (d *boltDB) Update(fn func(tx Tx) error) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		return fn(&boltTx{tx})
	})
}

func (d *boltDB) View(fn func(tx Tx) error) error {
	return d.db.View(func(tx *bolt.Tx) error {
		return fn(&boltTx{tx})
	})
}

type boltTx struct {
	tx *bolt.Tx
}

func (t *boltTx) Insert(bucketName, key []byte, data interface{}) error {
	dataBytes, err := encodeData(data)
	if err != nil {
		return err
	}

	if bucket, err := t.tx.Bucket(bucketName).CreateBucket(key); err != nil {
		if err == bolt.ErrTxNotWritable {
			return ErrTxNotWritable.New("Could not insert record")
		} else {
			return ErrKeyAlreadyExists.New("Could not insert record")
		}
	} else {
		if nextInt, err := bucket.NextSequence(); err != nil {
			return err
		} else {
			var numericKey [8]byte
			binary.BigEndian.PutUint64(numericKey[:], nextInt)
			if err = bucket.Put(numericKey[:], dataBytes); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *boltTx) AddIndex(indexBucket, index, key []byte) error {
	bucket := t.tx.Bucket(indexBucket)
	if bucket.Get(index) != nil {
		return ErrKeyAlreadyExists.New("Could not add index")
	} else if err := bucket.Put(index, key); err == bolt.ErrTxNotWritable {
		return ErrTxNotWritable.New("Could not add index")
	} else {
		return err
	}
}

func (t *boltTx) Update(bucketName, key []byte, data interface{}) error {
	dataBytes, err := encodeData(data)
	if err != nil {
		return err
	}

	bucket := t.tx.Bucket(bucketName).Bucket(key)
	if bucket == nil {
		return ErrKeyDoesNotExist.New("Could not update nonexistent record")
	}
	if nextInt, err := bucket.NextSequence(); err != nil {
		if err == bolt.ErrTxNotWritable {
			return ErrTxNotWritable.New("Could not update record")
		} else {
			return err
		}
	} else {
		var numericKey [8]byte
		binary.BigEndian.PutUint64(numericKey[:], nextInt)
		if err = bucket.Put(numericKey[:], dataBytes); err != nil {
			return err
		}
	}
	return nil
}

func (t *boltTx) NextSequenceForBucket(bucket []byte) (uint64, error) {
	b := t.tx.Bucket(bucket)
	n, err := b.NextSequence()
	b.Put([]byte("SEQ-BUG-FIX"), []byte{})
	return n, err
}

func (t *boltTx) Get(bucketName, key []byte, data interface{}) error {
	bucket := t.tx.Bucket(bucketName).Bucket(key)
	if bucket == nil {
		return ErrKeyDoesNotExist.New("Could not get record")
	}
	_, buf := bucket.Cursor().Last()
	if data == nil {
		return ErrKeyDoesNotExist.New("Could not get record")
	}
	return decodeData(buf, data)
}

func (t *boltTx) GetByIndex(indexBucket, dataBucket, index []byte, data interface{}) error {
	return t.Get(dataBucket, t.tx.Bucket(indexBucket).Get(index), data)
}

func (t *boltTx) CreateBucketIfNotExists(name []byte) error {
	_, err := t.tx.CreateBucketIfNotExists(name)
	if err == bolt.ErrTxNotWritable {
		return ErrTxNotWritable.New("Could not create bucket")
	} else if err != nil {
		return ErrGeneric.New("Could not create bucket")
	}
	return err
}
