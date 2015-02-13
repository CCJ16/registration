package boltorm

import (
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
		So(ErrTxNotWritable.Contains(err), ShouldBeTrue)
	}
}

func sharedTests(db DB) func() {
	return func() {
		Convey("Attempting to create buckets in a read only transaction", func() {
			err := db.View(func(tx Tx) error {
				return tx.CreateBucketIfNotExists(bucket1)
			})
			Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
		})
		Convey("With buckets created", func() {
			So(db.Update(func(tx Tx) error {
				return tx.CreateBucketIfNotExists(bucket1)
			}), ShouldBeNil)
			So(db.Update(func(tx Tx) error {
				return tx.CreateBucketIfNotExists(bucket2)
			}), ShouldBeNil)
			Convey("Next Sequence works (starting from 1)", func() {
				var n uint64
				So(db.Update(func(tx Tx) error {
					var err error
					n, err = tx.NextSequenceForBucket(bucket1)
					return err
				}), ShouldBeNil)
				So(n, ShouldEqual, 1)
				So(db.Update(func(tx Tx) error {
					var err error
					n, err = tx.NextSequenceForBucket(bucket1)
					return err
				}), ShouldBeNil)
				So(n, ShouldEqual, 2)
				So(db.Update(func(tx Tx) error {
					var err error
					n, err = tx.NextSequenceForBucket(bucket1)
					return err
				}), ShouldBeNil)
				So(n, ShouldEqual, 3)
				Convey("Which is separate from a different bucket", func() {
					var n uint64
					So(db.Update(func(tx Tx) error {
						var err error
						n, err = tx.NextSequenceForBucket(bucket2)
						return err
					}), ShouldBeNil)
					So(n, ShouldEqual, 1)
					So(db.Update(func(tx Tx) error {
						var err error
						n, err = tx.NextSequenceForBucket(bucket2)
						return err
					}), ShouldBeNil)
					So(n, ShouldEqual, 2)
					So(db.Update(func(tx Tx) error {
						var err error
						n, err = tx.NextSequenceForBucket(bucket2)
						return err
					}), ShouldBeNil)
					So(n, ShouldEqual, 3)
				})
			})

			Convey("Attempting to insert records in a read only transaction", func() {
				data := testData{5}
				err := db.View(func(tx Tx) error {
					return tx.Insert(bucket1, []byte("KeyA"), &data)
				})
				Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
			})
			Convey("Attempting to add indexes in a read only transaction", func() {
				err := db.View(func(tx Tx) error {
					return tx.AddIndex(bucket2, bucket1, []byte("KeyA"))
				})
				Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
			})

			Convey("When inserting a record", func() {
				data := testData{5}
				err := db.Update(func(tx Tx) error {
					return tx.Insert(bucket1, []byte("KeyA"), &data)
				})
				Convey("It should succeed", func() {
					So(err, ShouldBeNil)
					Convey("And trying to reinsert the data", func() {
						err := db.Update(func(tx Tx) error {
							return tx.Insert(bucket1, []byte("KeyA"), &data)
						})
						Convey("Should fail with an already inserted error", func() {
							So(ErrKeyAlreadyExists.Contains(err), ShouldBeTrue)
						})
					})
					Convey("And fetching that record", func() {
						newData := testData{}
						err := db.View(func(tx Tx) error {
							return tx.Get(bucket1, []byte("KeyA"), &newData)
						})
						Convey("Should work without error", func() {
							So(err, ShouldBeNil)
							Convey("And have the original data", func() {
								So(newData, ShouldResemble, testData{5})
							})
						})
					})
					Convey("And storing an index", func() {
						err := db.Update(func(tx Tx) error {
							return tx.AddIndex(bucket2, []byte("IndexA"), []byte("KeyA"))
						})
						Convey("Should succeed", func() {
							So(err, ShouldBeNil)
							Convey("And fetching the record through the index works", func() {
								newData := testData{}
								err := db.View(func(tx Tx) error {
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
								err := db.Update(func(tx Tx) error {
									return tx.AddIndex(bucket2, []byte("IndexA"), []byte("KeyA"))
								})
								Convey("Should fail with a key already exists error", func() {
									So(ErrKeyAlreadyExists.Contains(err), ShouldBeTrue)
								})
							})
						})
					})
					Convey("Fetching the record through a nonexistent index should fail", func() {
						newData := testData{}
						err := db.View(func(tx Tx) error {
							return tx.GetByIndex(bucket2, bucket1, []byte("IndexA"), &newData)
						})
						Convey("By throwing a key not existing error", func() {
							So(ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
						})
					})
					Convey("When updating an existing record", func() {
						data := testData{7}
						err := db.Update(func(tx Tx) error {
							return tx.Update(bucket1, []byte("KeyA"), &data)
						})
						Convey("It should succeed", func() {
							So(err, ShouldBeNil)
							Convey("And fetching that record", func() {
								newData := testData{}
								err := db.View(func(tx Tx) error {
									return tx.Get(bucket1, []byte("KeyA"), &newData)
								})
								Convey("Should work without error", func() {
									So(err, ShouldBeNil)
									Convey("And have the new data", func() {
										So(newData, ShouldResemble, testData{7})
									})
								})
								Convey("When updating an existing record", func() {
									data := testData{9}
									err := db.Update(func(tx Tx) error {
										return tx.Update(bucket1, []byte("KeyA"), &data)
									})
									Convey("It should succeed", func() {
										So(err, ShouldBeNil)
										Convey("And fetching that record", func() {
											newData := testData{}
											err := db.View(func(tx Tx) error {
												return tx.Get(bucket1, []byte("KeyA"), &newData)
											})
											Convey("Should work without error", func() {
												So(err, ShouldBeNil)
												Convey("And have the new data", func() {
													So(newData, ShouldResemble, testData{9})
												})
											})
										})
									})
								})
							})
						})
					})
					Convey("And attempting to update the record in a read only transaction", func() {
						data := testData{5}
						err := db.View(func(tx Tx) error {
							return tx.Update(bucket1, []byte("KeyA"), &data)
						})
						Convey("Should fail with transaction is read only error", txReadOnlyTest(err))
					})
					Convey("And fetching all records should work", func() {
						var list []*testData
						err := db.View(func(tx Tx) error {
							if d, err := tx.GetAll(bucket1, &testData{}); err != nil {
								return err
							} else {
								list = d.([]*testData)
								return nil
							}
						})
						So(err, ShouldBeNil)
						Convey("With only my one record found", func() {
							So(list, ShouldResemble, []*testData{{5}})
						})
					})
					Convey("And with an extra record", func() {
						data := testData{6}
						err := db.Update(func(tx Tx) error {
							return tx.Insert(bucket1, []byte("KeyB"), &data)
						})
						So(err, ShouldBeNil)
						Convey("A successful fetching of all records", func() {
							var list []*testData
							err := db.View(func(tx Tx) error {
								if d, err := tx.GetAll(bucket1, &testData{}); err != nil {
									return err
								} else {
									list = d.([]*testData)
									return nil
								}
							})
							So(err, ShouldBeNil)
							Convey("Should return both", func() {
								So(list, ShouldResemble, []*testData{{5}, {6}})
							})
						})
					})
				})
			})
			Convey("Fetching a nonexistent record", func() {
				newData := testData{}
				err := db.View(func(tx Tx) error {
					return tx.Get(bucket1, []byte("KeyA"), &newData)
				})
				Convey("Should fail with a key does not exist error", func() {
					So(ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
				})
			})
			Convey("Updating a nonexistent record", func() {
				newData := testData{5}
				err := db.Update(func(tx Tx) error {
					return tx.Update(bucket1, []byte("KeyA"), &newData)
				})
				Convey("Should fail with a key does not exist error", func() {
					So(ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
				})
			})
			Convey("And storing an index to a nonexistent key", func() {
				err := db.Update(func(tx Tx) error {
					return tx.AddIndex(bucket2, []byte("IndexA"), []byte("KeyA"))
				})
				Convey("Should succeed", func() {
					So(err, ShouldBeNil)
					Convey("And fetching the record through the index should fail", func() {
						newData := testData{}
						err := db.View(func(tx Tx) error {
							return tx.GetByIndex(bucket2, bucket1, []byte("IndexA"), &newData)
						})
						Convey("By throwing a key not existing error", func() {
							So(ErrKeyDoesNotExist.Contains(err), ShouldBeTrue)
						})
					})
				})
			})
		})
	}
}
