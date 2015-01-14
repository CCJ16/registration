package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

type testPreRegDb struct {
	entries []GroupPreRegistration
}

func (db *testPreRegDb) CreateRecord(in *GroupPreRegistration) error {
	if err := in.PrepareForInsert(); err != nil {
		return err
	}
	db.entries = append(db.entries, *in)
	return nil
}

func (d *testPreRegDb) GetRecord(securityKey string) (rec *GroupPreRegistration, err error) {
	for _, rec := range d.entries {
		if rec.SecurityKey == securityKey {
			return &rec, nil
		}
	}
	return nil, RecordDoesNotExist.New("Record with given key (%s) does not exist.", securityKey)
}

func TestPreRegCreateRequest(t *testing.T) {
	Convey("Starting with a Group Pre Registration handler", t, func() {
		goodRecord := GroupPreRegistration{
			PackName:             "Pack A",
			GroupName:            "Test Group",
			Council:              "1st Testingway",
			EmailApprovalGivenAt: time.Now(),
		}
		goodRecordBody := bytes.Buffer{}
		if bytes, err := json.Marshal(goodRecord); err != nil {
			t.Fatal(err)
		} else {
			goodRecordBody.Write(bytes)
		}

		prdb := &testPreRegDb{}
		router := mux.NewRouter()
		prh := NewGroupPreRegistrationHandler(router, prdb)

		Convey("When given a good record", func() {
			r, err := http.NewRequest("POST", "http://localhost:8080/preregistration", &goodRecordBody)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			Convey("Should receive back a 201 status code", func() {
				So(w.Code, ShouldEqual, 201)
			})

			Convey("Should receive back a valid location for the resource", func() {
				newRec := GroupPreRegistration{}
				So(json.NewDecoder(w.Body).Decode(&newRec), ShouldBeNil)
				So(w.HeaderMap["Location"], ShouldResemble, []string{"/preregistration/" + string(newRec.SecurityKey)})
				location := w.HeaderMap["Location"][0]
				Convey("And fetching that location will return the same record back", func() {
					r, err := http.NewRequest("GET", "http://localhost:8080"+location, nil)
					if err != nil {
						t.Fatal(err)
					}
					w := httptest.NewRecorder()

					router.ServeHTTP(w, r)

					Convey("with a 200 status code", func() {
						So(w.Code, ShouldEqual, 200)
					})

					Convey("With the same object back", func() {
						getRec := GroupPreRegistration{}
						So(json.NewDecoder(w.Body).Decode(&getRec), ShouldBeNil)
						So(getRec, ShouldResemble, newRec)
					})
				})
			})

			Convey("Should get back the same object with a new security key", func() {
				newRec := GroupPreRegistration{}
				So(json.NewDecoder(w.Body).Decode(&newRec), ShouldBeNil)
				So(newRec.SecurityKey, ShouldNotBeEmpty)
				goodRecord.SecurityKey = newRec.SecurityKey
				So(newRec.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, goodRecord.EmailApprovalGivenAt)
				newRec.EmailApprovalGivenAt = goodRecord.EmailApprovalGivenAt;
				So(newRec, ShouldResemble, goodRecord)
				Convey("With a matching security key to the database", func() {
					So(prdb.entries[0].SecurityKey, ShouldEqual, newRec.SecurityKey)
				})
			})

			Convey("Should be inserted into the database", func() {
				So(len(prdb.entries), ShouldEqual, 1)
				goodRecord.SecurityKey = prdb.entries[0].SecurityKey
				goodRecord.ValidationToken = prdb.entries[0].ValidationToken
				So(prdb.entries[0].EmailApprovalGivenAt, ShouldHappenWithin, time.Second, goodRecord.EmailApprovalGivenAt)
				prdb.entries[0].EmailApprovalGivenAt = goodRecord.EmailApprovalGivenAt;
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
				So(len(rec.SecurityKey), ShouldEqual, keyLength/3*4) // Length of key once converted to base64
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

				Convey("And fetchable through the api", func() {
					fetchedRec, err := prdb.GetRecord(rec.SecurityKey)
					So(err, ShouldBeNil)
					So(*fetchedRec, ShouldResemble, rec)
				})
			})

			Convey("Should fail with an error if done again with a security key already set", func() {
				So(RecordAlreadyPrepared.Contains(prdb.CreateRecord(&rec)), ShouldBeTrue)
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
