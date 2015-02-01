package boltorm

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMemoryDb(t *testing.T) {
	Convey("With a memory DB", t, func() {
		db := NewMemoryDB()
		Convey("the standard tests work", sharedTests(db))
	})
}
