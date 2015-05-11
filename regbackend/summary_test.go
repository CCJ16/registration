package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/CCJ16/registration/regbackend/boltorm"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSummaryPackEndPoint(t *testing.T) {
	Convey("Starting with a summary api handler", t, func() {

		db := boltorm.NewMemoryDB()
		invDb, err := NewInvoiceDb(db)
		So(err, ShouldBeNil)
		prdb, err := NewPreRegBoltDb(db, invDb)
		So(err, ShouldBeNil)
		router := mux.NewRouter()
		sh := NewSummaryHandler(router, prdb)
		So(sh, ShouldNotBeNil)

		Convey("Requesting the pack summary should give a 200 output with a body", func() {
			r, err := http.NewRequest("GET", "http://localhost:8080/summary/pack", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			Convey("And the returned data should be an appropriate JSON structure", func() {
				type output struct {
					YouthCount  int
					LeaderCount int
				}
				outputVar := output{}
				err := json.Unmarshal(w.Body.Bytes(), &outputVar)
				So(err, ShouldBeNil)

				Convey("With 0 youth", func() {
					So(outputVar.YouthCount, ShouldEqual, 0)
				})
				Convey("With 0 leaders", func() {
					So(outputVar.LeaderCount, ShouldEqual, 0)
				})
			})
		})

		Convey("With a filled in database", func() {
			rec := GroupPreRegistration{
				PackName:           "Pack A",
				GroupName:          "Test Group",
				Council:            "1st Testingway",
				ContactLeaderEmail: "testemail@example.test",
				EstimatedYouth:     5,
				EstimatedLeaders:   3,
			}
			prdb.CreateRecord(&rec)
			rec = GroupPreRegistration{
				PackName:           "Pack B",
				GroupName:          "Test Group",
				Council:            "1st Testingway",
				ContactLeaderEmail: "testemail@example.test",
				EstimatedYouth:     12,
				EstimatedLeaders:   6,
			}
			prdb.CreateRecord(&rec)
			Convey("The api endpoint should give a 200 output", func() {
				r, err := http.NewRequest("GET", "http://localhost:8080/summary/pack", nil)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, 200)
				Convey("And the returned data should be an appropriate JSON structure", func() {
					type output struct {
						YouthCount  int
						LeaderCount int
					}
					outputVar := output{}
					err := json.Unmarshal(w.Body.Bytes(), &outputVar)
					So(err, ShouldBeNil)

					Convey("With 17 youth", func() {
						So(outputVar.YouthCount, ShouldEqual, 17)
					})
					Convey("With 9 leaders", func() {
						So(outputVar.LeaderCount, ShouldEqual, 9)
					})
				})
			})
		})
	})
}
