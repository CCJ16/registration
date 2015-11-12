package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/CCJ16/registration/regbackend/boltorm"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	. "github.com/smartystreets/goconvey/convey"
)

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

		db := boltorm.NewMemoryDB()
		invDb, err := NewInvoiceDb(db)
		So(err, ShouldBeNil)

		config := &configType{}
		config.General.EnableGroupReg = true

		prdb, err := NewPreRegBoltDb(db, config, invDb)
		So(err, ShouldBeNil)
		router := mux.NewRouter()
		testEmailSender := &testEmailSender{}
		ces := NewConfirmationEmailService("examplesite.com", "no-reply@examplesender.com", "no-reply", "info@infoexample.com", testEmailSender, prdb)
		prh := NewGroupPreRegistrationHandler(router, config, prdb, nil, ces)

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

				Convey("Should be inserted into the database", func() {
					dbRec, err := prdb.GetRecord(newRec.SecurityKey)
					Convey("As we didn't get an error", func() {
						So(err, ShouldBeNil)
						Convey("And the data matches", func() {
							So(newRec.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, dbRec.EmailApprovalGivenAt)
							dbRec.EmailApprovalGivenAt = newRec.EmailApprovalGivenAt
							So(newRec.ValidatedOn, ShouldHappenWithin, time.Second, dbRec.ValidatedOn)
							dbRec.ValidatedOn = newRec.ValidatedOn
							dbRec.ValidationToken = newRec.ValidationToken
							dbRec.EmailConfirmationSent = newRec.EmailConfirmationSent

							So(*dbRec, ShouldResemble, newRec)
						})
						Convey("And sending a request to confirm the correct email address with the correct code", func() {
							buf := bytes.NewReader([]byte(dbRec.ValidationToken))
							r, err := http.NewRequest("PUT", "http://localhost:8080/confirmpreregistration?email="+goodRecord.ContactLeaderEmail, buf)
							So(err, ShouldBeNil)
							w := httptest.NewRecorder()

							router.ServeHTTP(w, r)
							Convey("Should return a 200 status code", func() {
								So(w.Code, ShouldEqual, 200)
								Convey("And set the contact is validated.", func() {
									dbRec, err := prdb.GetRecord(newRec.SecurityKey)
									So(err, ShouldBeNil)
									So(dbRec.ValidatedOn, ShouldHappenWithin, time.Second*1, time.Now())
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
									dbRec, err := prdb.GetRecord(newRec.SecurityKey)
									So(err, ShouldBeNil)
									So(dbRec.ValidatedOn, ShouldHappenWithin, time.Second*0, time.Time{})
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
									dbRec, err := prdb.GetRecord(newRec.SecurityKey)
									So(err, ShouldBeNil)
									So(dbRec.ValidatedOn, ShouldHappenWithin, time.Second*0, time.Time{})
								})
							})
						})
					})
				})
				Convey("And requesting an invoice should be error free", func() {
					r, err := http.NewRequest("GET", "http://localhost:8080"+w.HeaderMap["Location"][0]+"/invoice", nil)
					if err != nil {
						t.Fatal(err)
					}
					w := httptest.NewRecorder()

					router.ServeHTTP(w, r)

					Convey("with a 200 status code", func() {
						So(w.Code, ShouldEqual, 200)
					})

					Convey("Should get a valid invoice back.", func() {
						inv := Invoice{}
						So(json.NewDecoder(w.Body).Decode(&inv), ShouldBeNil)
						Convey("And match the database", func() {
							dbInv, err := prdb.CreateInvoiceIfNotExists(&newRec)
							So(err, ShouldBeNil)
							So(inv.Created, ShouldHappenWithin, 0*time.Second, dbInv.Created)
							inv.Created = dbInv.Created
							So(inv, ShouldResemble, *dbInv)
						})
					})
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
		})

		Convey("When inserted with registrations closed", func() {
			config.General.EnableGroupReg = false

			r, err := http.NewRequest("POST", "http://localhost:8080/preregistration", &goodRecordBody)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			Convey("Should receive back a 403 status code", func() {
				So(w.Code, ShouldEqual, 403)
			})
		})
	})
}

func CompareList(actual []*GroupPreRegistration, expected map[string]*GroupPreRegistration) {
	So(len(expected), ShouldEqual, len(actual))
	for _, gpr := range actual {
		expectedValue := expected[gpr.SecurityKey]
		So(gpr.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, expectedValue.EmailApprovalGivenAt)
		gpr.EmailApprovalGivenAt = expectedValue.EmailApprovalGivenAt
		So(gpr.ValidatedOn, ShouldHappenWithin, time.Second, expectedValue.ValidatedOn)
		gpr.ValidatedOn = expectedValue.ValidatedOn
		gpr.ValidationToken = expectedValue.ValidationToken
		gpr.EmailConfirmationSent = expectedValue.EmailConfirmationSent

		So(gpr, ShouldResemble, expectedValue)

		delete(expected, gpr.SecurityKey)
	}
}

func TestPreRegWaitingListHandler(t *testing.T) {
	Convey("Starting with a valid handler with 2 groups registered and 3 waiting", t, func() {
		db := boltorm.NewMemoryDB()
		invDb, err := NewInvoiceDb(db)
		So(err, ShouldBeNil)
		config := &configType{}
		prdb, err := NewPreRegBoltDb(db, config, invDb)
		So(err, ShouldBeNil)
		router := mux.NewRouter()
		testEmailSender := &testEmailSender{}
		ces := NewConfirmationEmailService("examplesite.com", "no-reply@examplesender.com", "no-reply", "info@infoexample.com", testEmailSender, prdb)
		store := sessions.NewCookieStore([]byte("A"))
		prh := NewGroupPreRegistrationHandler(router, config, prdb, &AuthenticationHandler{config, store}, ces)

		reg1 := &GroupPreRegistration{
			PackName:           "Pack A",
			GroupName:          "1st Testingway",
			Council:            "Council rock",
			ContactLeaderEmail: "testemail@example.com",
		}
		So(prdb.CreateRecord(reg1), ShouldBeNil)

		reg2 := &GroupPreRegistration{
			PackName:           "Pack B",
			GroupName:          "1st Testingway",
			Council:            "Council rock",
			ContactLeaderEmail: "testemail2@example.com",
		}
		So(prdb.CreateRecord(reg2), ShouldBeNil)

		config.General.EnableWaitingList = true

		wait1 := &GroupPreRegistration{
			PackName:           "Pack A",
			GroupName:          "1st Testingway",
			Council:            "Waiters rock",
			ContactLeaderEmail: "testemail3@example.com",
		}
		So(prdb.CreateRecord(wait1), ShouldBeNil)

		wait2 := &GroupPreRegistration{
			PackName:           "Pack B",
			GroupName:          "1st Testingway",
			Council:            "Waiters rock",
			ContactLeaderEmail: "testemail4@example.com",
		}
		So(prdb.CreateRecord(wait2), ShouldBeNil)

		wait3 := &GroupPreRegistration{
			PackName:           "Pack C",
			GroupName:          "1st Testingway",
			Council:            "Waiters rock",
			ContactLeaderEmail: "testemail5@example.com",
		}
		So(prdb.CreateRecord(wait3), ShouldBeNil)

		Convey("Fetching the default record list", func() {
			r, err := http.NewRequest("GET", "http://localhost:8080/preregistration", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			prh.GetList(w, r)

			Convey("Should receive back a 200 code", func() {
				So(w.Code, ShouldEqual, 200)
				Convey("With a valid json list", func() {
					recs := []*GroupPreRegistration{}
					So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
					Convey("With all records", func() {
						CompareList(recs, map[string]*GroupPreRegistration{
							reg1.SecurityKey:  reg1,
							reg2.SecurityKey:  reg2,
							wait1.SecurityKey: wait1,
							wait2.SecurityKey: wait2,
							wait3.SecurityKey: wait3,
						})
					})
				})
			})
		})

		Convey("Fetching the all inclusive record list", func() {
			r, err := http.NewRequest("GET", "http://localhost:8080/preregistration?select=all", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			prh.GetList(w, r)

			Convey("Should receive back a 200 code", func() {
				So(w.Code, ShouldEqual, 200)
				Convey("With a valid json list", func() {
					recs := []*GroupPreRegistration{}
					So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
					Convey("With all records", func() {
						CompareList(recs, map[string]*GroupPreRegistration{
							reg1.SecurityKey:  reg1,
							reg2.SecurityKey:  reg2,
							wait1.SecurityKey: wait1,
							wait2.SecurityKey: wait2,
							wait3.SecurityKey: wait3,
						})
					})
				})
			})
		})

		Convey("Fetching only the registered record list", func() {
			r, err := http.NewRequest("GET", "http://localhost:8080/preregistration?select=registered", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			prh.GetList(w, r)

			Convey("Should receive back a 200 code", func() {
				So(w.Code, ShouldEqual, 200)
				Convey("With a valid json list", func() {
					recs := []*GroupPreRegistration{}
					So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
					Convey("With all records", func() {
						CompareList(recs, map[string]*GroupPreRegistration{
							reg1.SecurityKey: reg1,
							reg2.SecurityKey: reg2,
						})
					})
				})
			})
		})

		Convey("Fetching only the waiting record list", func() {
			r, err := http.NewRequest("GET", "http://localhost:8080/preregistration?select=waiting", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			prh.GetList(w, r)

			Convey("Should receive back a 200 code", func() {
				So(w.Code, ShouldEqual, 200)
				Convey("With a valid json list with three records", func() {
					recs := []*GroupPreRegistrationInWaitingList{}
					So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
					So(len(recs), ShouldEqual, 3)
					Convey("With all records in order", func() {
						Convey("For the first record", func() {
							Convey("Should have position 1", func() {
								So(recs[0].WaitingListPos, ShouldEqual, 1)
								Convey("With the correct data", func() {
									So(recs[0].GroupPreRegistration.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, wait1.EmailApprovalGivenAt)
									recs[0].GroupPreRegistration.EmailApprovalGivenAt = wait1.EmailApprovalGivenAt
									So(recs[0].GroupPreRegistration.ValidatedOn, ShouldHappenWithin, time.Second, wait1.ValidatedOn)
									recs[0].GroupPreRegistration.ValidatedOn = wait1.ValidatedOn
									recs[0].GroupPreRegistration.ValidationToken = wait1.ValidationToken
									recs[0].GroupPreRegistration.EmailConfirmationSent = wait1.EmailConfirmationSent

									So(recs[0].GroupPreRegistration, ShouldResemble, wait1)
								})
							})
						})
						Convey("For the second record", func() {
							Convey("Should have position 2", func() {
								So(recs[1].WaitingListPos, ShouldEqual, 2)
								Convey("With the correct data", func() {
									So(recs[1].GroupPreRegistration.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, wait2.EmailApprovalGivenAt)
									recs[1].GroupPreRegistration.EmailApprovalGivenAt = wait2.EmailApprovalGivenAt
									So(recs[1].GroupPreRegistration.ValidatedOn, ShouldHappenWithin, time.Second, wait2.ValidatedOn)
									recs[1].GroupPreRegistration.ValidatedOn = wait2.ValidatedOn
									recs[1].GroupPreRegistration.ValidationToken = wait2.ValidationToken
									recs[1].GroupPreRegistration.EmailConfirmationSent = wait2.EmailConfirmationSent

									So(recs[1].GroupPreRegistration, ShouldResemble, wait2)
								})
							})
						})
						Convey("For the third record", func() {
							Convey("Should have position 3", func() {
								So(recs[2].WaitingListPos, ShouldEqual, 3)
								Convey("With the correct data", func() {
									So(recs[2].GroupPreRegistration.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, wait3.EmailApprovalGivenAt)
									recs[2].GroupPreRegistration.EmailApprovalGivenAt = wait3.EmailApprovalGivenAt
									So(recs[2].GroupPreRegistration.ValidatedOn, ShouldHappenWithin, time.Second, wait3.ValidatedOn)
									recs[2].GroupPreRegistration.ValidatedOn = wait3.ValidatedOn
									recs[2].GroupPreRegistration.ValidationToken = wait3.ValidationToken
									recs[2].GroupPreRegistration.EmailConfirmationSent = wait3.EmailConfirmationSent

									So(recs[2].GroupPreRegistration, ShouldResemble, wait3)
								})
							})
						})
					})
				})
			})
		})
		Convey("And promoting the second record to a full registration", func() {
			r, err := http.NewRequest("POST", "http://localhost:8080/preregistration/"+wait2.SecurityKey+"/promote", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			Convey("While logged in", func() {
				// Fake being logged in
				sess, err := sessions.GetRegistry(r).Get(store, globalSessionName)
				So(err, ShouldBeNil)
				So(sess, ShouldNotBeNil)
				sess.Values[authStatusLoggedIn] = true

				router.ServeHTTP(w, r)
				Convey("Should receive back a 200 code", func() {
					So(w.Code, ShouldEqual, 200)
					wait2.IsOnWaitingList = false // This will be updated by above.
					Convey("And trying again on the same record", func() {
						r, err := http.NewRequest("POST", "http://localhost:8080/preregistration/"+wait2.SecurityKey+"/promote", nil)
						if err != nil {
							t.Fatal(err)
						}
						w := httptest.NewRecorder()

						// Fake being logged in
						sess, err := sessions.GetRegistry(r).Get(store, globalSessionName)
						So(err, ShouldBeNil)
						So(sess, ShouldNotBeNil)
						sess.Values[authStatusLoggedIn] = true

						router.ServeHTTP(w, r)
						Convey("Should receive back a 400 code", func() {
							So(w.Code, ShouldEqual, 400)
						})
					})

					Convey("Fetching only the registered record list", func() {
						r, err := http.NewRequest("GET", "http://localhost:8080/preregistration?select=registered", nil)
						if err != nil {
							t.Fatal(err)
						}
						w := httptest.NewRecorder()

						prh.GetList(w, r)

						Convey("Should receive back a 200 code", func() {
							So(w.Code, ShouldEqual, 200)
							Convey("With a valid json list", func() {
								recs := []*GroupPreRegistration{}
								So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
								Convey("With all records including the new waitlist", func() {
									CompareList(recs, map[string]*GroupPreRegistration{
										reg1.SecurityKey:  reg1,
										reg2.SecurityKey:  reg2,
										wait2.SecurityKey: wait2,
									})
								})
							})
						})
					})

					Convey("Fetching only the waiting record list", func() {
						r, err := http.NewRequest("GET", "http://localhost:8080/preregistration?select=waiting", nil)
						if err != nil {
							t.Fatal(err)
						}
						w := httptest.NewRecorder()

						prh.GetList(w, r)

						Convey("Should receive back a 200 code", func() {
							So(w.Code, ShouldEqual, 200)
							Convey("With a valid json list with two records, missing the lost waiting list", func() {
								recs := []*GroupPreRegistrationInWaitingList{}
								So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
								So(len(recs), ShouldEqual, 2)
								Convey("With all records in order", func() {
									Convey("For the first record", func() {
										Convey("Should have position 1", func() {
											So(recs[0].WaitingListPos, ShouldEqual, 1)
											Convey("With the correct data", func() {
												So(recs[0].GroupPreRegistration.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, wait1.EmailApprovalGivenAt)
												recs[0].GroupPreRegistration.EmailApprovalGivenAt = wait1.EmailApprovalGivenAt
												So(recs[0].GroupPreRegistration.ValidatedOn, ShouldHappenWithin, time.Second, wait1.ValidatedOn)
												recs[0].GroupPreRegistration.ValidatedOn = wait1.ValidatedOn
												recs[0].GroupPreRegistration.ValidationToken = wait1.ValidationToken
												recs[0].GroupPreRegistration.EmailConfirmationSent = wait1.EmailConfirmationSent

												So(recs[0].GroupPreRegistration, ShouldResemble, wait1)
											})
										})
									})
									Convey("For the second record", func() {
										Convey("Should have position 2", func() {
											So(recs[1].WaitingListPos, ShouldEqual, 2)
											Convey("With the correct data", func() {
												So(recs[1].GroupPreRegistration.EmailApprovalGivenAt, ShouldHappenWithin, time.Second, wait3.EmailApprovalGivenAt)
												recs[1].GroupPreRegistration.EmailApprovalGivenAt = wait3.EmailApprovalGivenAt
												So(recs[1].GroupPreRegistration.ValidatedOn, ShouldHappenWithin, time.Second, wait3.ValidatedOn)
												recs[1].GroupPreRegistration.ValidatedOn = wait3.ValidatedOn
												recs[1].GroupPreRegistration.ValidationToken = wait3.ValidationToken
												recs[1].GroupPreRegistration.EmailConfirmationSent = wait3.EmailConfirmationSent

												So(recs[1].GroupPreRegistration, ShouldResemble, wait3)
											})
										})
									})
								})
							})
						})
					})
				})
			})
			Convey("While not logged in", func() {
				router.ServeHTTP(w, r)
				Convey("Should receive back a 403 code", func() {
					So(w.Code, ShouldEqual, 403)

					Convey("And fetching only the registered record list", func() {
						r, err := http.NewRequest("GET", "http://localhost:8080/preregistration?select=registered", nil)
						if err != nil {
							t.Fatal(err)
						}
						w := httptest.NewRecorder()

						prh.GetList(w, r)

						Convey("Should receive back a 200 code", func() {
							So(w.Code, ShouldEqual, 200)
							Convey("With a valid json list", func() {
								recs := []*GroupPreRegistration{}
								So(json.Unmarshal(w.Body.Bytes(), &recs), ShouldBeNil)
								Convey("With all records the same, as no promotion happened.", func() {
									CompareList(recs, map[string]*GroupPreRegistration{
										reg1.SecurityKey: reg1,
										reg2.SecurityKey: reg2,
									})
								})
							})
						})
					})
				})
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

		config := &configType{}
		dbOrm := boltorm.NewBoltDB(db)
		invDb, err := NewInvoiceDb(dbOrm)
		So(err, ShouldBeNil)

		prdb, err := NewPreRegBoltDb(dbOrm, config, invDb)
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

			Convey("Should be registered, not waiting", func() {
				So(rec.IsOnWaitingList, ShouldBeFalse)
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

			Convey("And requesting an invoice should be error free", func() {
				inv, err := prdb.CreateInvoiceIfNotExists(&rec)
				So(err, ShouldBeNil)
				Convey("With a matching set of invoice ids", func() {
					So(inv.ID, ShouldEqual, rec.InvoiceID)
				})
				Convey("And the invoice should have", func() {
					Convey("A valid to line", func() {
						So(inv.To, ShouldEqual, "1st Testingway of Council rock (Pack A)")
					})
					Convey("With the single deposit item", func() {
						So(inv.LineItems, ShouldResemble, []InvoiceItem{{"Pre-registration deposit", 25000, 1}})
					})
				})
				Convey("And refetching the record should keep the same contents", func() {
					inv2, err := prdb.CreateInvoiceIfNotExists(&rec)
					So(err, ShouldBeNil)
					So(inv2, ShouldResemble, inv)
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

		Convey("Inserting a group without a pack name", func() {
			rec := GroupPreRegistration{
				GroupName:          "1st Testingway",
				Council:            "Council rock",
				ContactLeaderEmail: "testemail@example.com",
			}
			So(prdb.CreateRecord(&rec), ShouldBeNil)

			Convey("Should still work with invoices", func() {
				inv, err := prdb.CreateInvoiceIfNotExists(&rec)
				So(err, ShouldBeNil)
				Convey("With a matching set of invoice ids", func() {
					So(inv.ID, ShouldEqual, rec.InvoiceID)
				})
				Convey("And the invoice should have", func() {
					Convey("A valid to line", func() {
						So(inv.To, ShouldEqual, "1st Testingway of Council rock")
					})
					Convey("With the single deposit item", func() {
						So(inv.LineItems, ShouldResemble, []InvoiceItem{{"Pre-registration deposit", 25000, 1}})
					})
				})
			})
		})

		Convey("And with the system in wait list mode", func() {
			config.General.EnableWaitingList = true
			Convey("Inserting a group", func() {
				rec := GroupPreRegistration{
					GroupName:          "1st Testingway",
					Council:            "Council rock",
					ContactLeaderEmail: "testemail@example.com",
				}
				So(prdb.CreateRecord(&rec), ShouldBeNil)

				Convey("Should insert them set in the waiting list", func() {
					So(rec.IsOnWaitingList, ShouldBeTrue)
					Convey("And put them in the waiting list index", func() {
						key, err := base64.URLEncoding.DecodeString(rec.SecurityKey)
						So(err, ShouldBeNil)

						So(db.View(func(tx *bolt.Tx) error {
							bucket := tx.Bucket(BOLT_GROUPEWAITINGLISTBUCKET)
							So(bucket.Stats().KeyN, ShouldEqual, 1)
							So(bucket.Get([]byte{0, 0, 0, 0, 0, 0, 0, 1}), ShouldResemble, key)
							return nil
						}), ShouldBeNil)
					})
				})
				Convey("Inserting a second group", func() {
					rec2 := GroupPreRegistration{
						GroupName:          "2st Testingway",
						Council:            "Council rock",
						ContactLeaderEmail: "testemail2@example.com",
					}
					So(prdb.CreateRecord(&rec2), ShouldBeNil)

					Convey("Should insert them set in the waiting list", func() {
						So(rec.IsOnWaitingList, ShouldBeTrue)
						Convey("And put them in the waiting list as the second group", func() {
							key, err := base64.URLEncoding.DecodeString(rec.SecurityKey)
							So(err, ShouldBeNil)
							key2, err := base64.URLEncoding.DecodeString(rec2.SecurityKey)
							So(err, ShouldBeNil)

							So(db.View(func(tx *bolt.Tx) error {
								bucket := tx.Bucket(BOLT_GROUPEWAITINGLISTBUCKET)
								So(bucket.Stats().KeyN, ShouldEqual, 2)
								So(bucket.Get([]byte{0, 0, 0, 0, 0, 0, 0, 1}), ShouldResemble, key)
								So(bucket.Get([]byte{0, 0, 0, 0, 0, 0, 0, 2}), ShouldResemble, key2)
								return nil
							}), ShouldBeNil)
						})
					})
				})

				Convey("Should not allow invoices", func() {
					inv, err := prdb.CreateInvoiceIfNotExists(&rec)
					So(inv, ShouldBeNil)
					Convey("With the appropriate error", func() {
						So(NoInvoiceOnWaitingList.Contains(err), ShouldBeTrue)
					})
				})
			})
		})

		Convey("Fetching a missing record", func() {
			So(config.General.EnableWaitingList, ShouldEqual, false)
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
