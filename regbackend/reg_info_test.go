package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	entries [][]GroupPreRegistration
}

func (db *testPreRegDb) CreateRecord(in *GroupPreRegistration) error {
	if err := in.PrepareForInsert(); err != nil {
		return err
	}
	for _, rec := range db.entries {
		if rec[len(rec)-1].OrganicKey() == in.OrganicKey() {
			return GroupAlreadyCreated.New("Group %s of %s, with pack name %s already exists", in.GroupName, in.Council, in.PackName)
		}
		if rec[len(rec)-1].ContactLeaderEmail == in.ContactLeaderEmail {
			return GroupAlreadyCreated.New("A previous group already registered with contact email address %s\n", in.ContactLeaderEmail)
		}
	}
	db.entries = append(db.entries, []GroupPreRegistration{*in})
	return nil
}

func (d *testPreRegDb) GetRecord(securityKey string) (rec *GroupPreRegistration, err error) {
	for _, rec := range d.entries {
		if rec[0].SecurityKey == securityKey {
			return &rec[len(rec)-1], nil
		}
	}
	return nil, RecordDoesNotExist.New("Record with given key (%s) does not exist.", securityKey)
}

func (d *testPreRegDb) VerifyEmail(email, token string) error {
	recLoc := -1
	for i, rec := range d.entries {
		if rec[len(rec)-1].ContactLeaderEmail == email {
			recLoc = i
		}
	}
	if recLoc == -1 {
		return BadVerificationToken.New("Failed to verify key")
	}
	recArray := d.entries[recLoc]
	rec := &recArray[len(recArray)-1]
	if rec.ValidationToken == token {
		rec.ValidatedOn = time.Now()
		return nil
	}
	return BadVerificationToken.New("Failed to verify key")
}

func (d *testPreRegDb) NoteConfirmationEmailSent(gpr *GroupPreRegistration) error {
	recLoc := -1
	for i, rec := range d.entries {
		if rec[0].SecurityKey == gpr.SecurityKey {
			recLoc = i
		}
	}
	if recLoc == -1 {
		return RecordDoesNotExist.New("Record with given key (%s) does not exist.", gpr.SecurityKey)
	}
	rec := d.entries[recLoc][len(d.entries[recLoc])-1]
	rec.EmailConfirmationSent = true
	d.entries[recLoc] = append(d.entries[recLoc], rec)
	gpr.EmailConfirmationSent = true
	return nil
}

func TestPreRegCreateRequest(t *testing.T) {
	Convey("Starting with a Group Pre Registration handler", t, func() {
		goodRecord := GroupPreRegistration{
			PackName:             "Pack A",
			GroupName:            "Test Group",
			Council:              "1st Testingway",
			ContactLeaderEmail:   "testemail@example.test",
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
		testEmailSender := &testEmailSender{}
		ces := NewConfirmationEmailService("examplesite.com", "no-reply@examplesender.com", "no-reply", "info@infoexample.com", testEmailSender, prdb)
		prh := NewGroupPreRegistrationHandler(router, prdb, ces)

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
				newRec.EmailApprovalGivenAt = goodRecord.EmailApprovalGivenAt
				So(newRec.ValidatedOn, ShouldHappenWithin, time.Second, goodRecord.ValidatedOn)
				newRec.ValidatedOn = goodRecord.ValidatedOn
				So(newRec, ShouldResemble, goodRecord)
				Convey("With a matching security key to the database", func() {
					So(prdb.entries[0][0].SecurityKey, ShouldEqual, newRec.SecurityKey)
				})
			})

			Convey("Should be inserted into the database", func() {
				So(len(prdb.entries), ShouldEqual, 1)
				So(len(prdb.entries[0]), ShouldEqual, 2)
				Convey("With the first record equal to second, modulo the email confirmation sent", func() {
					tmpRec := prdb.entries[0][0]
					tmpRec.EmailConfirmationSent = prdb.entries[0][1].EmailConfirmationSent
					So(prdb.entries[0][1], ShouldResemble, tmpRec)
				})
				Convey("With the second record confirming the email sent", func() {
					So(prdb.entries[0][1].EmailConfirmationSent, ShouldResemble, true)
				})
				Convey("And the first record matching the request, modulo the keys, and funky time business", func() {
					goodRecord.SecurityKey = prdb.entries[0][0].SecurityKey
					goodRecord.ValidationToken = prdb.entries[0][0].ValidationToken
					So(prdb.entries[0][0].EmailApprovalGivenAt, ShouldHappenWithin, time.Second, goodRecord.EmailApprovalGivenAt)
					prdb.entries[0][0].EmailApprovalGivenAt = goodRecord.EmailApprovalGivenAt
					So(prdb.entries[0][0].ValidatedOn, ShouldHappenWithin, time.Second, goodRecord.ValidatedOn)
					prdb.entries[0][0].ValidatedOn = goodRecord.ValidatedOn
					So(prdb.entries[0][0], ShouldResemble, goodRecord)
				})
			})

			Convey("And attempting to re create the same group", func() {
				Convey("Should fail with an error if done again with the same group set", func() {
					goodRecord.ContactLeaderEmail = "newemail@example.test"
					goodRecordBody := bytes.Buffer{}
					if bytes, err := json.Marshal(goodRecord); err != nil {
						t.Fatal(err)
					} else {
						goodRecordBody.Write(bytes)
					}
					r, err := http.NewRequest("POST", "http://localhost:8080/preregistration", &goodRecordBody)
					if err != nil {
						t.Fatal(err)
					}
					w := httptest.NewRecorder()

					router.ServeHTTP(w, r)

					Convey("With a 400 status code", func() {
						So(w.Code, ShouldEqual, 400)
						Convey("And error message", func() {
							So(string(w.Body.Bytes()), ShouldEqual, fmt.Sprintf("Group already registered: Group %s of %s, with pack name %s already exists\n", goodRecord.GroupName, goodRecord.Council, goodRecord.PackName))
						})
					})
				})

				Convey("Should fail with an error if done again with the same email set", func() {
					goodRecord.Council = "Test Council 2"
					goodRecordBody := bytes.Buffer{}
					if bytes, err := json.Marshal(goodRecord); err != nil {
						t.Fatal(err)
					} else {
						goodRecordBody.Write(bytes)
					}
					r, err := http.NewRequest("POST", "http://localhost:8080/preregistration", &goodRecordBody)
					if err != nil {
						t.Fatal(err)
					}
					w := httptest.NewRecorder()

					router.ServeHTTP(w, r)

					Convey("With a 400 status code", func() {
						So(w.Code, ShouldEqual, 400)
						Convey("And error message", func() {
							So(string(w.Body.Bytes()), ShouldEqual, fmt.Sprintf("Group already registered: A previous group already registered with contact email address %s\n", goodRecord.ContactLeaderEmail))
						})
					})
				})
			})

			Convey("And sending a request to confirm the correct email address with the correct code", func() {
				buf := bytes.NewReader([]byte(prdb.entries[0][1].ValidationToken))
				r, err := http.NewRequest("PUT", "http://localhost:8080/confirmpreregistration?email="+goodRecord.ContactLeaderEmail, buf)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, r)
				Convey("Should return a 200 status code", func() {
					So(w.Code, ShouldEqual, 200)
					Convey("And set the contact is validated.", func() {
						So(prdb.entries[0][len(prdb.entries[0])-1].ValidatedOn, ShouldHappenWithin, 5*time.Second, time.Now())
					})
				})
			})
			Convey("And sending a request to confirm the correct email address with the wrong code", func() {
				buf := bytes.NewReader([]byte("BadToken"))
				r, err := http.NewRequest("PUT", "http://localhost:8080/confirmpreregistration?email="+goodRecord.ContactLeaderEmail, buf)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, r)
				Convey("Should return a 200 status code", func() {
					So(w.Code, ShouldEqual, 400)
					Convey("And set the contact is not validated.", func() {
						So(prdb.entries[0][len(prdb.entries[0])-1].ValidatedOn, ShouldHappenWithin, time.Second*0, time.Time{})
					})
				})
			})
			Convey("And sending a request to confirm the wrong email address with a code", func() {
				buf := bytes.NewReader([]byte("BadToken"))
				r, err := http.NewRequest("PUT", "http://localhost:8080/confirmpreregistration?email=abademail@invalid", buf)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, r)
				Convey("Should return a 200 status code", func() {
					So(w.Code, ShouldEqual, 400)
					Convey("And set the contact is not validated.", func() {
						So(prdb.entries[0][len(prdb.entries[0])-1].ValidatedOn, ShouldHappenWithin, time.Second*0, time.Time{})
					})
				})
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
					bucket := tx.Bucket(BOLT_GROUPBUCKET).Bucket(rec.Key())
					So(bucket, ShouldNotBeNil)
					size := 0
					var data []byte
					So(bucket.ForEach(func(k, v []byte) error {
						size++
						So(v, ShouldNotBeNil)
						data = v
						return nil
					}), ShouldBeNil)
					So(size, ShouldEqual, 1)
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

			Convey("Noting a successful email conversion (with slightly modified data)", func() {
				rec.ContactLeaderFirstName = "New Name"
				err := prdb.NoteConfirmationEmailSent(&rec)
				Convey("Without error", func() {
					So(err, ShouldBeNil)
				})
				Convey("Should mark my copy as having been sent", func() {
					So(rec.EmailConfirmationSent, ShouldEqual, true)
					Convey("Without other modifications", func() {
						So(rec.ContactLeaderFirstName, ShouldEqual, "New Name")
					})
				})
				Convey("And update the database", func() {
					record := GroupPreRegistration{}
					So(db.View(func(tx *bolt.Tx) error {
						bucket := tx.Bucket(BOLT_GROUPBUCKET).Bucket(rec.Key())
						So(bucket, ShouldNotBeNil)
						size := 0
						var data []byte
						So(bucket.ForEach(func(k, v []byte) error {
							size++
							So(v, ShouldNotBeNil)
							data = v
							return nil
						}), ShouldBeNil)
						Convey("By adding a new version", func() {
							So(size, ShouldEqual, 2)
						})
						decoder := gob.NewDecoder(bytes.NewReader(data))
						return decoder.Decode(&record)
					}), ShouldBeNil)
					Convey("With the new value", func() {
						So(record.EmailConfirmationSent, ShouldEqual, true)
					})
					Convey("And not update other fields", func() {
						So(record.ContactLeaderFirstName, ShouldNotEqual, rec.ContactLeaderFirstName)
					})
					Convey("And a retry of the operation is silently ignored", func() {
						rec.EmailConfirmationSent = false
						err := prdb.NoteConfirmationEmailSent(&rec)
						Convey("Thus no error", func() {
							So(err, ShouldBeNil)
						})
						Convey("And there should only be the two records", func() {
							size := 0
							So(db.View(func(tx *bolt.Tx) error {
								bucket := tx.Bucket(BOLT_GROUPBUCKET).Bucket(rec.Key())
								So(bucket, ShouldNotBeNil)
								So(bucket.ForEach(func(k, v []byte) error {
									size++
									So(v, ShouldNotBeNil)
									return nil
								}), ShouldBeNil)
								return nil
							}), ShouldBeNil)
							So(size, ShouldEqual, 2)
						})
					})
				})
			})

			Convey("And verifying a valid token", func() {
				err := prdb.VerifyEmail(rec.ContactLeaderEmail, rec.ValidationToken)
				Convey("Should complete without error", func() {
					So(err, ShouldBeNil)
				})
				Convey("And update the database", func() {
					record := GroupPreRegistration{}
					So(db.View(func(tx *bolt.Tx) error {
						bucket := tx.Bucket(BOLT_GROUPBUCKET).Bucket(rec.Key())
						So(bucket, ShouldNotBeNil)
						size := 0
						var data []byte
						So(bucket.ForEach(func(k, v []byte) error {
							size++
							So(v, ShouldNotBeNil)
							data = v
							return nil
						}), ShouldBeNil)
						Convey("By adding a new version", func() {
							So(size, ShouldEqual, 2)
						})
						decoder := gob.NewDecoder(bytes.NewReader(data))
						return decoder.Decode(&record)
					}), ShouldBeNil)
					Convey("With the new value", func() {
						So(record.ValidatedOn, ShouldHappenWithin, time.Second*5, time.Now())
					})
					Convey("And a retry of the operation is silently ignored", func() {
						err := prdb.VerifyEmail(rec.ContactLeaderEmail, rec.ValidationToken)
						Convey("Thus no error", func() {
							So(err, ShouldBeNil)
						})
						Convey("And there should only be the two records", func() {
							size := 0
							So(db.View(func(tx *bolt.Tx) error {
								bucket := tx.Bucket(BOLT_GROUPBUCKET).Bucket(rec.Key())
								So(bucket, ShouldNotBeNil)
								So(bucket.ForEach(func(k, v []byte) error {
									size++
									So(v, ShouldNotBeNil)
									return nil
								}), ShouldBeNil)
								return nil
							}), ShouldBeNil)
							So(size, ShouldEqual, 2)
						})
					})
				})
			})

			Convey("And verifying an invalid token", func() {
				err := prdb.VerifyEmail(rec.ContactLeaderEmail, "BadToken")
				Convey("Should complete with appropriate error", func() {
					So(BadVerificationToken.Contains(err), ShouldEqual, true)
				})
			})

			Convey("And verifying an email not in the database", func() {
				err := prdb.VerifyEmail("test@invalid", rec.ValidationToken)
				Convey("Should complete with appropriate error (not revealing email error)", func() {
					So(BadVerificationToken.Contains(err), ShouldEqual, true)
				})
			})
		})
		Convey("Fetching a missing record", func() {
			fetchedRec, err := prdb.GetRecord("aaaa")
			Convey("Should return a nil record", func() {
				So(fetchedRec, ShouldBeNil)
			})
			Convey("Should return a record not found error", func() {
				So(RecordDoesNotExist.Contains(err), ShouldBeTrue)
			})
		})
	})
}
