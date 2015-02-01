package main

import (
	"encoding/binary"
	"time"

	"github.com/CCJ16/registration/regbackend/boltorm"
)

type Invoice struct {
	Id        uint64
	To        string
	LineItems []InvoiceItem
	Created   time.Time
}

type InvoiceItem struct {
	Description string
	UnitPrice   int64
	Count       int64
}

type InvoiceDb interface {
	NewInvoice(in *Invoice, tx boltorm.Tx) error
	GetInvoice(invoiceId uint64, tx boltorm.Tx) (*Invoice, error)
}

type invoiceDb struct {
}

var (
	BOLT_INVOICEBUCKET = []byte("BUCKET_INVOICES")
)

func NewInvoiceDb(db boltorm.DB) (InvoiceDb, error) {
	if err := db.Update(func(tx boltorm.Tx) error {
		return tx.CreateBucketIfNotExists(BOLT_INVOICEBUCKET)
	}); err != nil {
		return nil, err
	}
	return &invoiceDb{}, nil
}

func (i *invoiceDb) NewInvoice(in *Invoice, tx boltorm.Tx) error {
	in.Id = 0
	for in.Id == 0 {
		id, err := tx.NextSequenceForBucket(BOLT_INVOICEBUCKET)
		if err != nil {
			return err
		}
		in.Id = id
	}
	in.Created = time.Now()
	var idBytes [8]byte
	binary.BigEndian.PutUint64(idBytes[:], in.Id)
	return tx.Insert(BOLT_INVOICEBUCKET, idBytes[:], in)
}

func (i *invoiceDb) GetInvoice(invoiceId uint64, tx boltorm.Tx) (inv *Invoice, err error) {
	inv = &Invoice{}
	var key [8]byte
	binary.BigEndian.PutUint64(key[:], invoiceId)
	if err = tx.Get(BOLT_INVOICEBUCKET, key[:], inv); err != nil {
		if boltorm.ErrKeyDoesNotExist.Contains(err) {
			return nil, RecordDoesNotExist.New("Could not find invoice")
		} else {
			return nil, err
		}
	}
	return inv, nil
}
