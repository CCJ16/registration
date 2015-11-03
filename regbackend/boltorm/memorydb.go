package boltorm

import (
	"bytes"
	"sort"
	"sync"
)

type bucketData struct {
	data map[string][][]byte
	seq  uint64
}

type memoryDB struct {
	buckets *map[string]*bucketData
	lock    sync.RWMutex
}

func NewMemoryDB() DB {
	buckets := make(map[string]*bucketData)
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
	buckets  *map[string]*bucketData
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

	if (*t.buckets)[string(bucket)].data[string(key)] != nil {
		return ErrKeyAlreadyExists.New("Could not insert record")
	} else {
		(*t.buckets)[string(bucket)].data[string(key)] = [][]byte{dataBytes}
	}
	return nil
}

func (t *memoryTx) AddIndex(indexBucket, index, key []byte) error {
	if !t.writable {
		return ErrTxNotWritable.New("Could not insert record")
	}

	if (*t.buckets)[string(indexBucket)].data[string(index)] != nil {
		return ErrKeyAlreadyExists.New("Could not insert index")
	} else {
		(*t.buckets)[string(indexBucket)].data[string(index)] = [][]byte{key}
	}
	return nil
}

func (t *memoryTx) Update(bucket, key []byte, data interface{}) error {
	if !t.writable {
		return ErrTxNotWritable.New("Could not insert record")
	}
	dataBytes, err := encodeData(data)
	if err != nil {
		return err
	}

	if (*t.buckets)[string(bucket)].data[string(key)] != nil {
		(*t.buckets)[string(bucket)].data[string(key)] = append((*t.buckets)[string(bucket)].data[string(key)], dataBytes)
	} else {
		return ErrKeyDoesNotExist.New("Could not update record")
	}
	return nil
}

func (t *memoryTx) NextSequenceForBucket(bucket []byte) (uint64, error) {
	b := (*t.buckets)[string(bucket)]
	b.seq++
	return b.seq, nil
}

func (t *memoryTx) Get(bucket, key []byte, data interface{}) error {
	dataBucket := (*t.buckets)[string(bucket)].data[string(key)]
	if dataBucket == nil {
		return ErrKeyDoesNotExist.New("Failed to get record")
	}
	bytes := dataBucket[len(dataBucket)-1]
	return decodeData(bytes, data)
}

type sorter struct {
	key  string
	data interface{}
}

type sorterSort []sorter

func (s *sorterSort) Len() int {
	return len(*s)
}

func (s *sorterSort) Less(i, j int) bool {
	return (*s)[i].key < (*s)[j].key
}

func (s *sorterSort) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

func (t *memoryTx) GetAll(bucketName []byte, dataType interface{}) (interface{}, error) {
	ret := makeSliceFor(dataType)
	sortSlice := sorterSort{}
	bucket := (*t.buckets)[string(bucketName)]
	for key, dataBucket := range bucket.data {
		bytes := dataBucket[len(dataBucket)-1]
		nextElement := makeNew(dataType)
		if err := decodeData(bytes, nextElement); err != nil {
			return nil, err
		}
		sortSlice = append(sortSlice, sorter{string(key), nextElement})
	}
	sort.Sort(&sortSlice)
	for _, elm := range sortSlice {
		ret = appendToSlice(ret, elm.data)
	}
	return ret, nil
}

func (t *memoryTx) GetByIndex(indexBucket, dataBucket, index []byte, data interface{}) error {
	indexData := (*t.buckets)[string(indexBucket)].data[string(index)]
	if indexData == nil {
		return ErrKeyDoesNotExist.New("Failed to get key of record")
	}
	key := indexData[0]
	dataBucketMap := (*t.buckets)[string(dataBucket)].data[string(key)]
	if dataBucketMap == nil {
		return ErrKeyDoesNotExist.New("Failed to get record")
	}
	bytes := dataBucketMap[len(dataBucketMap)-1]
	return decodeData(bytes, data)
}

func (t *memoryTx) CreateBucketIfNotExists(name []byte) error {
	if t.writable {
		(*t.buckets)[string(name)] = &bucketData{
			data: make(map[string][][]byte),
			seq:  0,
		}
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

func (t *memoryTx) GetAllByIndex(indexBucket, dataBucket []byte, dataType interface{}) (interface{}, error) {
	ret := makeSliceFor(dataType)
	sortSlice := sorterSort{}
	iBucket := (*t.buckets)[string(indexBucket)]
	dBucket := (*t.buckets)[string(dataBucket)]
	for index, keyA := range iBucket.data {
		key := keyA[0]

		dataBucket := dBucket.data[string(key)]
		bytes := dataBucket[len(dataBucket)-1]

		nextElement := makeNew(dataType)
		if err := decodeData(bytes, nextElement); err != nil {
			return nil, err
		}
		sortSlice = append(sortSlice, sorter{index, nextElement})
	}
	sort.Sort(&sortSlice)
	for _, elm := range sortSlice {
		ret = appendToSlice(ret, elm.data)
	}
	return ret, nil
}

func (t *memoryTx) RemoveKeyFromIndex(indexBucket, key []byte) error {
	iBucket := (*t.buckets)[string(indexBucket)]
	for index, keyA := range iBucket.data {
		curKey := keyA[0]

		if bytes.Compare(curKey, key) == 0 {
			delete(iBucket.data, index)
		}
	}
	return nil
}
