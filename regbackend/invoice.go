package main

import (
	"encoding/binary"
	"time"

	"github.com/CCJ16/registration/regbackend/boltorm"
)

type Invoice struct {
	ID        uint64        `json:"id"`
	To        string        `json:"to"`
	LineItems []InvoiceItem `json:"lineItems"`
	Created   time.Time     `json:"created"`
}

type InvoiceItem struct {
	Description string `json:"description"`
	UnitPrice   int64  `json:"unitPrice"`
	Count       int64  `json:"count"`
}

type InvoiceDb interface {
	NewInvoice(in *Invoice, tx boltorm.Tx) error
	GetInvoice(invoiceID uint64, tx boltorm.Tx) (*Invoice, error)
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
	in.ID = 0
	for in.ID == 0 {
		id, err := tx.NextSequenceForBucket(BOLT_INVOICEBUCKET)
		if err != nil {
			return err
		}
		in.ID = id
	}
	in.Created = time.Now()
	var idBytes [8]byte
	binary.BigEndian.PutUint64(idBytes[:], in.ID)
	return tx.Insert(BOLT_INVOICEBUCKET, idBytes[:], in)
}

func (i *invoiceDb) GetInvoice(invoiceID uint64, tx boltorm.Tx) (inv *Invoice, err error) {
	inv = &Invoice{}
	var key [8]byte
	binary.BigEndian.PutUint64(key[:], invoiceID)
	if err = tx.Get(BOLT_INVOICEBUCKET, key[:], inv); err != nil {
		if boltorm.ErrKeyDoesNotExist.Contains(err) {
			return nil, RecordDoesNotExist.New("Could not find invoice")
		} else {
			return nil, err
		}
	}
	return inv, nil
}
