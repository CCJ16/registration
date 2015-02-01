package boltorm

import (
	"sync"
)

type memoryDB struct {
	buckets *map[string]map[string][][]byte
	lock    sync.RWMutex
}

func NewMemoryDB() DB {
	buckets := make(map[string]map[string][][]byte)
	return &memoryDB{
		buckets: &buckets,
	}
}

func (m *memoryDB) Update(fn func(tx Tx) error) error {
	m.lock.Lock()
	tx := &memoryTx{m, m.buckets, true, true}
	defer tx.rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.commit()
}

func (m *memoryDB) View(fn func(tx Tx) error) error {
	m.lock.Lock()
	tx := &memoryTx{m, m.buckets, true, false}
	defer tx.rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.rollback()
}

type memoryTx struct {
	m        *memoryDB
	buckets  *map[string]map[string][][]byte
	valid    bool
	writable bool
}

func (t *memoryTx) Insert(bucket, key []byte, data interface{}) error {
	if !t.writable {
		return ErrTxNotWritable.New("Could not insert record")
	}
	dataBytes, err := encodeData(data)
	if err != nil {
		return err
	}

	if (*t.buckets)[string(bucket)][string(key)] != nil {
		return ErrKeyAlreadyExists.New("Could not insert record")
	} else {
		(*t.buckets)[string(bucket)][string(key)] = [][]byte{dataBytes}
	}
	return nil
}

func (t *memoryTx) Get(bucket, key []byte, data interface{}) error {
	dataBucket := (*t.buckets)[string(bucket)][string(key)]
	if dataBucket == nil {
		return ErrKeyDoesNotExist.New("Failed to get record")
	}
	bytes := dataBucket[len(dataBucket)-1]
	return decodeData(bytes, data)
}

func (t *memoryTx) CreateBucketIfNotExists(name []byte) error {
	if t.writable {
		(*t.buckets)[string(name)] = make(map[string][][]byte)
		return nil
	} else {
		return ErrTxNotWritable.New("Could not create bucket")
	}
}

func (t *memoryTx) commit() error {
	t.m.buckets = t.buckets
	if t.valid {
		t.m.lock.Unlock()
		t.valid = false
	}
	return nil
}

func (t *memoryTx) rollback() error {
	if t.valid {
		t.m.lock.Unlock()
		t.valid = false
	}
	return nil
}
