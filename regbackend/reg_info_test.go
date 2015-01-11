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
	goodRecord := GroupPreRegistration{}
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

	Convey("Should return 201 and insert for a good record", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:8080/prereg", &goodRecordBody)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()

		prh.Create(w, r)

		if w.Code != 201 {
			t.Errorf("Did not receive a 201 response for record creation, got: %v", w.Code)
		}

		if len(prdb.entries) != 1 || prdb.entries[0] != goodRecord {
			t.Errorf("Failed to insert record!")
		}
	})
	Convey("Should return 400 for invalid json", t, func() {
		r, err := http.NewRequest("POST", "http://localhost:8080/prereg", &bytes.Buffer{})
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()

		prh.Create(w, r)

		if w.Code != 400 {
			t.Errorf("Did not receive a 400 for empty request: %v", w.Code)
		}
	})
}
