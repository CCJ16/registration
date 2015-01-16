package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	apiR := r.PathPrefix("/api/").Subrouter()

	db, err := bolt.Open("records.bolt", 0600, &bolt.Options{Timeout: 1})
	if err != nil {
		log.Fatalf("Failed to open bolt database, err: %s", err)
	}

	gprdb, err := NewPreRegBoltDb(db)
	if err != nil {
		log.Fatalf("Failed to get group preregistration database started, err %s", err)
	}

	ces := NewConfirmationEmailService("registration.cubjamboree.ca", "no-reply@cubjamboree.ca", "info@cubjamboree.ca", NewLocalMailder("localhost:25"), gprdb)

	NewGroupPreRegistrationHandler(apiR, gprdb, ces)

	http.Handle("/api/", r)
	http.Handle("/", http.FileServer(http.Dir("../app")))
	panic(http.ListenAndServe(":8080", nil))
}
