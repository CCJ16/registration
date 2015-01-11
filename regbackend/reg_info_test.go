package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"


	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/boltdb/bolt"

	. "github.com/smartystreets/goconvey/convey"
)

type testPreRegDb struct {
	entries []GroupPreRegistration
}

func (db *testPreRegDb) CreateRecord(in *GroupPreRegistration) error {
	db.entries = append(db.entries, *in)
	return nil
}

func TestPreRegCreateRequest(t *testing.T) {
	Convey("Starting with a Group Pre Registration handler", t, func() {
		goodRecord := GroupPreRegistration{
			PackName:  "Pack A",
			GroupName: "Test Group",
			Council:   "1st Testingway",
		}
		goodRecordBody := bytes.Buffer{}
		if bytes, err := json.Marshal(goodRecord); err != nil {
			t.Fatal(err)
		} else {
			goodRecordBody.Write(bytes)
		}

		prdb := &testPreRegDb{}
		prh := PreRegHandler{
			db: prdb,
		}

		Convey("When given a good record", func() {
			r, err := http.NewRequest("POST", "http://localhost:8080/prereg", &goodRecordBody)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			prh.Create(w, r)

			Convey("Should receive back a 201 status code", func() {
				So(w.Code, ShouldEqual, 201)
			})

			Convey("Should be inserted into the database", func() {
				So(len(prdb.entries), ShouldEqual, 1)
				So(prdb.entries[0], ShouldResemble, goodRecord)
			})
		})

		Convey("When given an empty body", func() {
			r, err := http.NewRequest("POST", "http://localhost:8080/prereg", &bytes.Buffer{})
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			prh.Create(w, r)

			Convey("Should receive back a 400 code", func() {
				So(w.Code, ShouldEqual, 400)
			})
			Convey("Shouldn't update the database", func() {
				So(prdb.entries, ShouldBeEmpty)
			})
		})
	})
}

func TestGroupPreRegDbInBolt(t *testing.T) {
	Convey("With a given bolt database", t, func() {
		file, err := ioutil.TempFile("", "")
		So(err, ShouldBeNil)
		Reset(func() {
			So(os.Remove(file.Name()), ShouldBeNil)
			file.Close()
		})

		db, err := bolt.Open(file.Name(), 0, nil)
		So(err, ShouldBeNil)
		Reset(func() {
			So(db.Close(), ShouldBeNil)
		})

		prdb, err := NewPreRegBoltDb(db)
		So(err, ShouldBeNil)

		Convey("Inserting a group", func() {
			rec := GroupPreRegistration{
				PackName:           "Pack A",
				GroupName:          "1st Testingway",
				Council:            "Council rock",
				ContactLeaderEmail: "testemail@example.com",
			}
			duprec := rec
			So(prdb.CreateRecord(&rec), ShouldBeNil)

			Convey("Should set a security key", func() {
				So(rec.SecurityKey, ShouldNotBeNil)
				So(len(rec.SecurityKey), ShouldEqual, 128)
			})

			Convey("Should make it available in bolt", func() {
				record := GroupPreRegistration{}
				So(db.View(func(tx *bolt.Tx) error {
					data := tx.Bucket(BOLT_GROUPBUCKET).Get(rec.Key())
					decoder := gob.NewDecoder(bytes.NewReader(data))
					return decoder.Decode(&record)
				}), ShouldBeNil)
				So(record, ShouldResemble, rec)

				Convey("And Organic key index exists", func() {
					var data []byte
					So(db.View(func(tx *bolt.Tx) error {
						data = tx.Bucket(BOLT_GROUPNAMEMAPBUCKET).Get([]byte(rec.OrganicKey()))
						return nil
					}), ShouldBeNil)
					So(data, ShouldResemble, rec.Key())
				})

				Convey("And email index exists", func() {
					var data []byte
					So(db.View(func(tx *bolt.Tx) error {
						data = tx.Bucket(BOLT_GROUPEMAILMAPBUCKET).Get([]byte(rec.ContactLeaderEmail))
						return nil
					}), ShouldBeNil)
					So(data, ShouldResemble, rec.Key())
				})
			})

			Convey("Should fail with an error if done again with the same security key", func() {
				So(GroupAlreadyCreated.Contains(prdb.CreateRecord(&rec)), ShouldBeTrue)
			})

			Convey("Should fail with an error if done again with the same group set", func() {
				So(GroupAlreadyCreated.Contains(prdb.CreateRecord(&duprec)), ShouldBeTrue)
			})

			Convey("Should fail with an error if done again with the same email set", func() {
				duprec.Council = "Other Council"
				So(GroupAlreadyCreated.Contains(prdb.CreateRecord(&duprec)), ShouldBeTrue)
			})
		})
	})
}
