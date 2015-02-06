package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CCJ16/registration/regbackend/boltorm"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
	goflagutils "github.com/spacemonkeygo/flagfile/utils"
	"github.com/yosssi/boltstore/reaper"
	"github.com/yosssi/boltstore/store"
)

var (
	SetupErrors = errors.NewClass("Error during setup")
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
	Integration         bool   `default:"false" usage:"Set when running an integration binary for testing."`
	Develop             bool   `default:"false" usage:"Set when running a binary for development."`
}

func init() {
	goflagutils.Setup("http", &httpConfig)
	goflagutils.Setup("email", &emailConfig)
	goflagutils.Setup("", &generalConfig)
}

type requestLogger struct {
	H http.Handler
}

type wWrapperLogger struct {
	code int
	http.ResponseWriter
}

func (w *wWrapperLogger) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (h *requestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	wrappedW := &wWrapperLogger{
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
		log.Panicf("Got error while copying database %s", err)
	}
}

type sessionSaver struct {
	h http.Handler
}

type wWrapperSession struct {
	req *http.Request
	http.ResponseWriter
	valid bool
}

func (w *wWrapperSession) WriteHeader(code int) {
	err := sessions.Save(w.req, w.ResponseWriter)
	if err != nil {
		http.Error(w.ResponseWriter, "Failed to save user session", http.StatusServiceUnavailable)
		log.Print("Failed to setup user session: ", err)
		w.valid = false
		return
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *wWrapperSession) Write(p []byte) (int, error) {
	if !w.valid {
		return len(p), nil
	} else {
		return w.ResponseWriter.Write(p)
	}
}

func (h *sessionSaver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.h.ServeHTTP(&wWrapperSession{r, w, true}, r)
}

func httpError(w http.ResponseWriter, err error) {
	http.Error(w, errhttp.GetErrorBody(err), errhttp.GetStatusCode(err, 500))
}

type httpRouter interface {
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func setupStandardHandlers(globalRouter httpRouter, db *bolt.DB) (http.Handler, chan<- struct{}, <-chan struct{}, error) {
	key := generalConfig.AccessToken
	if key == "" {
		var random [32]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, nil, nil, SetupErrors.New("During startup, failed to get entropy", err)
		}
		key = base64.URLEncoding.EncodeToString(random[:])
		log.Print("DB Token: ", key)
	}

	r := mux.NewRouter()
	ormDb := boltorm.NewBoltDB(db)
	apiR := r.PathPrefix("/api/").Subrouter()
	boltStore, err := store.New(db, store.Config{
		SessionOptions: sessions.Options{
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 30,
			Secure:   !(generalConfig.Integration || generalConfig.Develop),
			HttpOnly: true,
		},
		DBOptions: store.Options{
			BucketName: []byte("SESSIONS_BUCKET"),
		},
	}, []byte(key))
	_ = boltStore
	if err != nil {
		return nil, nil, nil, SetupErrors.New("Failed to setup session data")
	}

	invDb, err := NewInvoiceDb(ormDb)
	if err != nil {
		return nil, nil, nil, SetupErrors.New("Failed to get invoice database started")
	}

	gprdb, err := NewPreRegBoltDb(ormDb, invDb)
	if err != nil {
		return nil, nil, nil, SetupErrors.New("Failed to get group preregistration database started", err)
	}

	ces := NewConfirmationEmailService(generalConfig.Domain, emailConfig.FromAddress, emailConfig.FromName, emailConfig.ContactEmail, NewLocalMailder(emailConfig.Server), gprdb)

	NewGroupPreRegistrationHandler(apiR, gprdb, ces)

	apiR.Handle("/grabdb", &grabDb{db}).Headers("X-My-Auth-Token", key).Methods("GET").Queries("key", key)

	globalRouter.Handle("/api/", r)
	otherFiles := http.FileServer(http.Dir(generalConfig.StaticFilesLocation))
	globalRouter.Handle("/app.css", otherFiles)
	globalRouter.Handle("/app.js", otherFiles)
	globalRouter.Handle("/components/", otherFiles)
	globalRouter.Handle("/views/", otherFiles)
	globalRouter.Handle("/bower_components/", otherFiles)
	indexLocation := generalConfig.StaticFilesLocation + "/index.html"
	globalRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexLocation)
	})
	quitC, doneC := reaper.Run(db, reaper.Options{BucketName: []byte("SESSIONS_BUCKET")})
	return &sessionSaver{globalRouter}, quitC, doneC, nil
}
