package boltorm_test

import (
	"github.com/CCJ16/registration/regbackend/boltorm"

	. "github.com/smartystreets/goconvey/convey"
)

type testData struct {
	I int
}

var (
	bucket1 = []byte("B1")
)

func txReadOnlyTest(err error) func() {
	return func() {
		So(boltorm.ErrTxNotWritable.Contains(err), ShouldBeTrue)
	}
}

func sharedTests(db boltorm.DB) func() {
	return func() {
		Convey("Attempting to create buckets in a read only transaction", func() {
			err := db.View(func(tx boltorm.Tx) error {
				return tx.CreateBucketIfNotExists(bucket1)
			})
			Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
		})
		Convey("Attempting to insert records in a read only transaction", func() {
			data := testData{5}
			err := db.View(func(tx boltorm.Tx) error {
				return tx.Insert(bucket1, []byte("KeyA"), &data)
			})
			Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
		})
		Convey("Attempting to update records in a read only transaction", func() {
			data := testData{5}
			err := db.View(func(tx boltorm.Tx) error {
				return tx.Update(bucket1, []byte("KeyA"), &data)
			})
			Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
		})
		Convey("With buckets created", func() {
			So(db.Update(func(tx boltorm.Tx) error {
				return tx.CreateBucketIfNotExists(bucket1)
			}), ShouldBeNil)

			Convey("When inserting a record", func() {
				data := testData{5}
				err := db.Update(func(tx boltorm.Tx) error {
					return tx.Insert(bucket1, []byte("KeyA"), &data)
				})
				Convey("It should succeed", func() {
					So(err, ShouldBeNil)
					Convey("And trying to reinsert the data", func() {
						err := db.Update(func(tx boltorm.Tx) error {
							return tx.Insert(bucket1, []byte("KeyA"), &data)
						})
						Convey("Should fail with an already inserted error", func() {
							So(boltorm.ErrKeyAlreadyExists.Contains(err), ShouldBeTrue)
						})
					})
					Convey("And fetching that record", func() {
						newData := testData{}
						err := db.View(func(tx boltorm.Tx) error {
							return tx.Get(bucket1, []byte("KeyA"), &newData)
						})
						Convey("Should work without error", func() {
							So(err, ShouldBeNil)
							Convey("And have the original data", func() {
								So(newData, ShouldResemble, data)
							})
						})
					})
					Convey("When updating an existing record", func() {
						data := testData{7}
						err := db.Update(func(tx boltorm.Tx) error {
							return tx.Update(bucket1, []byte("KeyA"), &data)
						})
						Convey("It should succeed", func() {
							So(err, ShouldBeNil)
							Convey("And fetching that record", func() {

								newData := testData{}
								err := db.View(func(tx boltorm.Tx) error {
									return tx.Get(bucket1, []byte("KeyA"), &newData)
								})
								Convey("Should work without error", func() {
									So(err, ShouldBeNil)
									Convey("And have the new data", func() {
										So(newData, ShouldResemble, data)
									})
								})
							})
						})
					})

				})
			})
			Convey("Fetching a nonexistant record", func() {
				newData := testData{}
				err := db.View(func(tx boltorm.Tx) error {
					return tx.Get(bucket1, []byte("KeyA"), &newData)
				})
				Convey("Should fail with a key does not exist error", func() {
					So(boltorm.ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
				})
			})
			Convey("Updating a nonexistant record", func() {
				newData := testData{5}
				err := db.Update(func(tx boltorm.Tx) error {
					return tx.Update(bucket1, []byte("KeyA"), &newData)
				})
				Convey("Should fail with a key does not exist error", func() {
					So(boltorm.ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
				})
			})

		})
	}
}
