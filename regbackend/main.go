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
	"github.com/spacemonkeygo/errors/errhttp"
	"github.com/spacemonkeygo/flagfile"
	goflagutils "github.com/spacemonkeygo/flagfile/utils"
)

var httpConfig struct {
	Listen string `default:":8080" usage:"Address for server to listen on"`
}

var emailConfig struct {
	FromAddress  string `default:"no-reply@invalid" usage:"From address for use in emails"`
	FromName     string `usage:"From name for use in emails"`
	ContactEmail string `default:"info@invalid" "usage:"Contact email address for use in emails"`
	Server       string `default:"localhost:25" usage:"Server to use for sending messages"`
}

var generalConfig struct {
	Domain              string `default:"invalid" usage:"Domain for use in emails, etc to link people to"`
	Database            string `default:"records.bolt" usage:"Location to store the database"`
	AccessToken         string `usage:"Token to access database.  Generated randomly and printed if not set"`
	StaticFilesLocation string `default:"../app" usage:"Location of static files for the site"`
}

func init() {
	goflagutils.Setup("http", &httpConfig)
	goflagutils.Setup("email", &emailConfig)
	goflagutils.Setup("", &generalConfig)
}

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

func httpError(w http.ResponseWriter, err error) {
	http.Error(w, errhttp.GetErrorBody(err), errhttp.GetStatusCode(err, 500))
}

func main() {
	flagfile.Load()

	r := mux.NewRouter()
	apiR := r.PathPrefix("/api/").Subrouter()

	db, err := bolt.Open(generalConfig.Database, 0600, &bolt.Options{Timeout: 1})
	if err != nil {
		log.Fatalf("Failed to open bolt database, err: %s", err)
	}

	gprdb, err := NewPreRegBoltDb(db)
	if err != nil {
		log.Fatalf("Failed to get group preregistration database started, err %s", err)
	}

	ces := NewConfirmationEmailService(generalConfig.Domain, emailConfig.FromAddress, emailConfig.FromName, emailConfig.ContactEmail, NewLocalMailder(emailConfig.Server), gprdb)

	NewGroupPreRegistrationHandler(apiR, gprdb, ces)

	key := generalConfig.AccessToken
	if key == "" {
		var random [32]byte
		if _, err := rand.Read(random[:]); err != nil {
			log.Fatalf("During startup, failed to get entropy with error %s", err)
		}
		key = base64.URLEncoding.EncodeToString(random[:])
		log.Print("DB Token: ", key)
	}
	apiR.Handle("/grabdb", &grabDb{db}).Headers("X-My-Auth-Token", key).Methods("GET").Queries("key", key)

	http.Handle("/api/", r)
	otherFiles := http.FileServer(http.Dir(generalConfig.StaticFilesLocation))
	http.Handle("/app.css", otherFiles)
	http.Handle("/app.js", otherFiles)
	http.Handle("/components/", otherFiles)
	http.Handle("/views/", otherFiles)
	http.Handle("/bower_components/", otherFiles)
	indexLocation := generalConfig.StaticFilesLocation + "/index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexLocation)
	})
	panic(http.ListenAndServe(httpConfig.Listen, handlers.CompressHandler(&requestLogger{http.DefaultServeMux})))
}
