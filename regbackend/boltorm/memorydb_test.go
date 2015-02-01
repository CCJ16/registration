package boltorm_test

import (
	"github.com/CCJ16/registration/regbackend/boltorm"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMemoryDb(t *testing.T) {
	Convey("With a memory DB", t, func() {
		db := boltorm.NewMemoryDB()
		Convey("the standard tests work", sharedTests(db))
	})
}
