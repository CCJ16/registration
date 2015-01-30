// +build !integration

package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/handlers"
	"github.com/spacemonkeygo/flagfile"
)

func main() {
	flagfile.Load()
	if generalConfig.Integration == true {
		log.Panic("Attempted to run in integration mode in a non-final binary!")
	}
	db, err := bolt.Open(generalConfig.Database, 0600, &bolt.Options{Timeout: 1})
	if err != nil {
		log.Fatalf("Failed to open bolt database, err: %s", err)
	}
	mux := http.NewServeMux()
	setupStandardHandlers(mux, db)
	panic(http.ListenAndServe(httpConfig.Listen, handlers.CompressHandler(&requestLogger{mux})))
}
