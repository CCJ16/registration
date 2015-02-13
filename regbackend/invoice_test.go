package main

import (
	"time"

	"github.com/CCJ16/registration/regbackend/boltorm"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestInvoiceStorage(t *testing.T) {
	Convey("With a valid invoice system", t, func() {
		db := boltorm.NewMemoryDB()
		invDb, err := NewInvoiceDb(db)
		So(err, ShouldBeNil)

		Convey("Requesting a new invoice should succeed", func() {
			invoice := Invoice{
				To: "Test group of Test Council",
				LineItems: []InvoiceItem{
					{
						Description: "Item 1",
						UnitPrice:   1034,
						Count:       3,
					},
					{
						Description: "Item 2",
						UnitPrice:   60,
						Count:       1000,
					},
				},
			}
			err := db.Update(func(tx boltorm.Tx) error {
				return invDb.NewInvoice(&invoice, tx)
			})
			So(err, ShouldBeNil)
			Convey("And requesting it again by id should succeed", func() {
				var dbInv *Invoice
				err := db.View(func(tx boltorm.Tx) error {
					var err error
					dbInv, err = invDb.GetInvoice(invoice.ID, tx)
					return err
				})
				So(err, ShouldBeNil)
				Convey("And the db invoice should be equivalent", func() {
					So(dbInv.Created, ShouldHappenWithin, 0*time.Second, invoice.Created)
					dbInv.Created = invoice.Created
					So(*dbInv, ShouldResemble, invoice)
				})
			})
		})
	})
}
