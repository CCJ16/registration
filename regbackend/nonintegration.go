// +build !integration

package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/handlers"
)

func main() {
	config := getConfig()

	if config.General.Integration == true {
		log.Panic("Attempted to run in integration mode in a non-final binary!")
	}
	if config.General.Develop != develop {
		log.Panic("Mismatch in development vs production build and configuration!")
	}
	db, err := bolt.Open(config.General.Database, 0600, &bolt.Options{Timeout: 1})
	if err != nil {
		log.Fatalf("Failed to open bolt database, err: %s", err)
	}
	mux := http.NewServeMux()
	realMux, _, _, err := setupStandardHandlers(mux, config, db)
	if err != nil {
		log.Fatalf("Failed to setup basic routing, err: %s", err)
	}
	panic(http.ListenAndServe(config.Http.Listen, handlers.CompressHandler(&requestLogger{realMux})))
}
