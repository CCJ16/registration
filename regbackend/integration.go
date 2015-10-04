// +build integration

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/handlers"
	"github.com/yosssi/boltstore/reaper"
)

var handlerForCookie = make(map[string]http.Handler)
var currentHandlerID int
var handlerForCookieLock sync.Mutex

type cleanupInfo struct {
	db    *bolt.DB
	quitC chan<- struct{}
	doneC <-chan struct{}
}

var dbs []*cleanupInfo
var dbsLock sync.Mutex
var wg sync.WaitGroup
var config *configType

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Kill, os.Interrupt)

	config = getConfig()
	if config.General.Integration == false {
		log.Panic("Attempted to run in normal mode in an integration binary!")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		c <- nil
		w.WriteHeader(200)
	})
	mux.HandleFunc("/", muxTest)
	mux.HandleFunc("/test_is_integration", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("true"))
	})

	l, err := net.Listen("tcp", config.Http.Listen)
	if err != nil {
		log.Fatalf("Failed to open tcp port, err %s", err)
	}
	server := http.Server{
		Handler:      handlers.CompressHandler(&requestLogger{mux}),
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
		ConnState:    ConnectionAccounting,
	}
	go func(c <-chan os.Signal) {
		<-c
		server.SetKeepAlivesEnabled(false)
		l.Close()
	}(c)
	if err := server.Serve(l); err != nil {
		if oe, ok := err.(*net.OpError); ok && oe.Op == "accept" && oe.Net == "tcp" && oe.Err.Error() == "use of closed network connection" {
			log.Print("Port nicely closed")
		} else {
			log.Fatalf("%#v", err)
		}
	}
	wg.Wait()
	dbsLock.Lock()
	defer dbsLock.Unlock()
	for _, dbStruct := range dbs {
		reaper.Quit(dbStruct.quitC, dbStruct.doneC)
		db := dbStruct.db
		path := db.Path()
		db.Close()
		err := os.Remove(path)
		if err != nil {
			log.Printf("Failed to remove database %s, err %s", path, err)
		}
	}
	log.Print("Cleanly done")
}

var conTrackIdle map[net.Conn]bool = make(map[net.Conn]bool)
var cTIM sync.RWMutex

func ConnectionAccounting(c net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		cTIM.Lock()
		defer cTIM.Unlock()
		conTrackIdle[c] = true
	case http.StateActive:
		cTIM.Lock()
		defer cTIM.Unlock()
		wg.Add(1)
		delete(conTrackIdle, c)
	case http.StateClosed:
		cTIM.RLock()
		if !conTrackIdle[c] {
			wg.Done()
		}
		cTIM.RUnlock()
		cTIM.Lock()
		defer cTIM.Unlock()
		delete(conTrackIdle, c)
	case http.StateIdle:
		wg.Done()
		cTIM.Lock()
		defer cTIM.Unlock()
		conTrackIdle[c] = true
	}
}

func muxTest(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler
	if cookie, err := r.Cookie("ClientID"); err != nil {
		handler = setupNewHandlers()
		handlerForCookieLock.Lock()
		newValue := currentHandlerID
		currentHandlerID++
		cookieValue := strconv.Itoa(newValue)
		handlerForCookie[cookieValue] = handler
		handlerForCookieLock.Unlock()
		cookie = &http.Cookie{
			Name:     "ClientID",
			Value:    cookieValue,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, cookie)
	} else {
		cookieValue := cookie.Value
		handlerForCookieLock.Lock()
		var ok bool
		handler, ok = handlerForCookie[cookieValue]
		handlerForCookieLock.Unlock()
		if !ok {
			http.SetCookie(w, &http.Cookie{
				Name:    "ClientID",
				MaxAge:  -1,
				Expires: time.Time{},
				Path:    "/",
			})
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusSeeOther)
			w.Write([]byte("Invalid cookie value!"))
			log.Print("Invalid cookie value!")
			return
		}
	}
	handler.ServeHTTP(w, r)
}

type dbWiper struct {
	db *bolt.DB
}

func (d *dbWiper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(k []byte, b *bolt.Bucket) error {
			return b.ForEach(func(k, v []byte) error {
				if v == nil {
					if err := b.DeleteBucket(k); err != nil {
						return err
					}
				} else {
					if err := b.Delete(k); err != nil {
						return err
					}
				}
				return nil
			})
		})
	})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type configHandler struct {
	config *configType
}

func (c *configHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This drains the connection pool
	cTIM.RLock()
	defer cTIM.RUnlock()
	wg.Add(-1)
	wg.Wait()
	wg.Add(1)

	if r.Method == "DELETE" {
		*c.config = *config
		w.WriteHeader(http.StatusOK)
	} else if r.Method == "GET" {
		if data, err := json.Marshal(c.config); err != nil {
			log.Print("Failed config object encode: ", err)
			httpError(w, err)
		} else {
			w.WriteHeader(http.StatusOK)
			if n, err := w.Write(data); err != nil || n != len(data) {
				log.Printf("Failed to write entire config object, got error %s, wrote %v", err, n)
			}
		}
	} else if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(c.config); err != nil {
			httpError(w, err)
			log.Print("Failed config object decode: ", err)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func setupNewHandlers() http.Handler {
	file, err := ioutil.TempFile("", "records")
	if err != nil {
		log.Fatalf("Failed to get temporary file, err: %s", err)
	}
	file.Close()
	db, err := bolt.Open(file.Name(), 0600, &bolt.Options{Timeout: 1})
	if err != nil {
		log.Fatalf("Failed to open bolt database, err: %s", err)
	}
	dbsLock.Lock()
	defer dbsLock.Unlock()
	tbc := cleanupInfo{db: db}
	dbs = append(dbs, &tbc)
	mux := http.NewServeMux()
	mux.Handle("/integration/wipe_database", &dbWiper{db})
	mux.HandleFunc("/integration/", http.NotFound)

	myConfig := *config
	mux.Handle("/integration/config", &configHandler{&myConfig})

	newHandler, quitC, doneC, err := setupStandardHandlers(mux, &myConfig, db)
	if err != nil {
		log.Fatalf("Failed to setup new standard handlers: %s", err)
	}
	tbc.quitC = quitC
	tbc.doneC = doneC
	return newHandler
}
