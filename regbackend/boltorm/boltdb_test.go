package boltorm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBoltDb(t *testing.T) {
	Convey("With a bolt DB", t, func() {
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

		Convey("the standard tests work", sharedTests(NewBoltDB(db)))
	})
}
