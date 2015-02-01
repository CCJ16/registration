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
	bucket2 = []byte("I1")
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
		Convey("Attempting to add indexes in a read only transaction", func() {
			err := db.View(func(tx boltorm.Tx) error {
				return tx.AddIndex(bucket2, bucket1, []byte("KeyA"))
			})
			Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
		})
		Convey("With buckets created", func() {
			So(db.Update(func(tx boltorm.Tx) error {
				return tx.CreateBucketIfNotExists(bucket1)
			}), ShouldBeNil)
			So(db.Update(func(tx boltorm.Tx) error {
				return tx.CreateBucketIfNotExists(bucket2)
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
					Convey("And storing an index", func() {
						err := db.Update(func(tx boltorm.Tx) error {
							return tx.AddIndex(bucket2, []byte("IndexA"), []byte("KeyA"))
						})
						Convey("Should succeed", func() {
							So(err, ShouldBeNil)
							Convey("And fetching the record through the index works", func() {
								newData := testData{}
								err := db.View(func(tx boltorm.Tx) error {
									return tx.GetByIndex(bucket2, bucket1, []byte("IndexA"), &newData)
								})
								Convey("Should work without error", func() {
									So(err, ShouldBeNil)
									Convey("And have the original data", func() {
										So(newData, ShouldResemble, data)
									})
								})
							})
							Convey("And storing the same index", func() {
								err := db.Update(func(tx boltorm.Tx) error {
									return tx.AddIndex(bucket2, []byte("IndexA"), []byte("KeyA"))
								})
								Convey("Should fail with a key already exists error", func() {
									So(boltorm.ErrKeyAlreadyExists.Contains(err), ShouldBeTrue)
								})
							})
						})
					})
					Convey("Fetching the record through a nonexistent index should fail", func() {
						newData := testData{}
						err := db.View(func(tx boltorm.Tx) error {
							return tx.GetByIndex(bucket2, bucket1, []byte("IndexA"), &newData)
						})
						Convey("By throwing a key not existing error", func() {
							So(boltorm.ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
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
			Convey("Fetching a nonexistent record", func() {
				newData := testData{}
				err := db.View(func(tx boltorm.Tx) error {
					return tx.Get(bucket1, []byte("KeyA"), &newData)
				})
				Convey("Should fail with a key does not exist error", func() {
					So(boltorm.ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
				})
			})
			Convey("Updating a nonexistent record", func() {
				newData := testData{5}
				err := db.Update(func(tx boltorm.Tx) error {
					return tx.Update(bucket1, []byte("KeyA"), &newData)
				})
				Convey("Should fail with a key does not exist error", func() {
					So(boltorm.ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
				})
			})
			Convey("And storing an index to a nonexistent key", func() {
				err := db.Update(func(tx boltorm.Tx) error {
					return tx.AddIndex(bucket2, []byte("IndexA"), []byte("KeyA"))
				})
				Convey("Should succeed", func() {
					So(err, ShouldBeNil)
					Convey("And fetching the record through the index should fail", func() {
						newData := testData{}
						err := db.View(func(tx boltorm.Tx) error {
							return tx.GetByIndex(bucket2, bucket1, []byte("IndexA"), &newData)
						})
						Convey("By throwing a key not existing error", func() {
							So(boltorm.ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
						})
					})
				})
			})
		})
	}
}
