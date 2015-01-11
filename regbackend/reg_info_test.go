package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testPreRegDb struct {
	entries []GroupPreRegistration
}

func (db *testPreRegDb) CreateRecord(in GroupPreRegistration) error {
	db.entries = append(db.entries, in)
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
