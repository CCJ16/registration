package boltorm

import (
	"github.com/spacemonkeygo/errors"
)

var (
	ErrGeneric          = errors.NewClass("Generic Error")
	ErrKeyAlreadyExists = ErrGeneric.NewClass("Key already exists")
	ErrKeyDoesNotExist  = ErrGeneric.NewClass("Key does not exist")
	ErrTxNotWritable    = ErrGeneric.NewClass("Transaction not writable")
)

type DB interface {
	Update(fn func(tx Tx) error) error
	View(fn func(tx Tx) error) error
}

type Tx interface {
	CreateBucketIfNotExists(name []byte) error
	Insert(bucket, key []byte, data interface{}) error
	Update(bucket, key []byte, data interface{}) error
	AddIndex(indexBucket, index, key []byte) error
	NextSequenceForBucket(bucket []byte) (uint64, error)

	Get(bucket, key []byte, data interface{}) error
	GetAll(bucket []byte, dataType interface{}) (interface{}, error)
	GetByIndex(indexBucket, dataBucket, index []byte, data interface{}) error
	GetAllByIndex(indexBucket, bucket []byte, dataType interface{}) (interface{}, error)
}
