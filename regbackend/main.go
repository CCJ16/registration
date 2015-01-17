package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type requestLogger struct {
	H http.Handler
}

type wWrapper struct {
	code int
	http.ResponseWriter
}

func (w *wWrapper) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (h *requestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	wrappedW := &wWrapper{
		ResponseWriter: w,
		code:           http.StatusOK, // Default code
	}
	h.H.ServeHTTP(wrappedW, r)
	duration := time.Now().Sub(start)
	log.Printf("Handled request for url %s, code %v, took %s seconds", r.URL, wrappedW.code, duration)
}

type grabDb struct {
	db *bolt.DB
}

func (h *grabDb) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.db.View(func(tx *bolt.Tx) error {
		w.Header()["Content-Length"] = []string{fmt.Sprint(tx.Size())}
		return tx.Copy(w)
	})
	if err != nil {
		log.Panicf("Got error copy database %s", err)
	}
}

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

	ces := NewConfirmationEmailService("registration.cubjamboree.ca", "no-reply@cubjamboree.ca", "CCJ16 Registration", "info@cubjamboree.ca", NewLocalMailder("localhost:25"), gprdb)

	NewGroupPreRegistrationHandler(apiR, gprdb, ces)

	var key string
	{
		var random [32]byte
		if _, err := rand.Read(random[:]); err != nil {
			log.Fatalf("During startup, failed to get entropy with error %s", err)
		}
		key = base64.URLEncoding.EncodeToString(random[:])
	}
	log.Print("DB Token: ", key)
	apiR.Handle("/grabdb", &grabDb{db}).Headers("X-My-Auth-Token", key).Methods("GET").Queries("key", key)

	http.Handle("/api/", r)
	otherFiles := http.FileServer(http.Dir("../app"))
	http.Handle("/app.css", otherFiles)
	http.Handle("/app.js", otherFiles)
	http.Handle("/components/", otherFiles)
	http.Handle("/views/", otherFiles)
	http.Handle("/bower_components/", otherFiles)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../app/index.html")
	})
	panic(http.ListenAndServe(":8080", handlers.CompressHandler(&requestLogger{http.DefaultServeMux})))
}
