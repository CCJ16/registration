package main

import (
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
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
		code: http.StatusOK, // Default code
	}
	h.H.ServeHTTP(wrappedW, r)
	duration := time.Now().Sub(start)
	log.Printf("Handled request for url %s, code %v, took %s seconds", r.URL, wrappedW.code, duration)
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

	ces := NewConfirmationEmailService("registration.cubjamboree.ca", "no-reply@cubjamboree.ca", "info@cubjamboree.ca", NewLocalMailder("localhost:25"), gprdb)

	NewGroupPreRegistrationHandler(apiR, gprdb, ces)

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
	panic(http.ListenAndServe(":8080", &requestLogger{http.DefaultServeMux}))
}
