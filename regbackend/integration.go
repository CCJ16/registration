// +build integration

package main

import (
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
	"github.com/spacemonkeygo/flagfile"
)

var handlerForCookie = make(map[string]http.Handler)
var currentHandlerId int
var handlerForCookieLock sync.Mutex
var dbs []*bolt.DB
var dbsLock sync.Mutex
var wg sync.WaitGroup

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Kill, os.Interrupt)

	flagfile.Load()
	if generalConfig.Integration == false {
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

	l, err := net.Listen("tcp", httpConfig.Listen)
	if err != nil {
		log.Fatalf("Failed to open tcp port, err %s", err)
	}
	go func(c <-chan os.Signal) {
		<-c
		l.Close()
	}(c)
	server := http.Server{
		Handler:      handlers.CompressHandler(&requestLogger{mux}),
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
		ConnState:    ConnectionAccounting,
	}
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
	for _, db := range dbs {
		path := db.Path()
		db.Close()
		err := os.Remove(path)
		if err != nil {
			log.Printf("Failed to remove database %s, err %s", path, err)
		}
	}
	log.Print("Cleanly done")
}

func ConnectionAccounting(_ net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		wg.Add(1)
	case http.StateClosed:
		wg.Done()
	}
}

func muxTest(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler
	if cookie, err := r.Cookie("ClientID"); err != nil {
		handler = setupNewHandlers()
		handlerForCookieLock.Lock()
		newValue := currentHandlerId
		currentHandlerId++
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
			});
		});
	})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
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
	dbs = append(dbs, db)
	dbsLock.Unlock()
	mux := http.NewServeMux()
	setupStandardHandlers(mux, db)
	mux.Handle("/integration/wipe_database", &dbWiper{db})
	mux.HandleFunc("/integration/", http.NotFound)
	return mux
}
