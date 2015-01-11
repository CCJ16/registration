package main

import (
	"crypto/rand"
	"fmt"
	"strings"

	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"bytes"
	"encoding/gob"
	"time"

	"github.com/boltdb/bolt"

	"github.com/spacemonkeygo/errors"
)

type PhoneNumber string

type Address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	Province   string `json:"province"`
	PostalCode string `json:"postalCode"`
}

type GroupPreRegistration struct {
	SecurityKey []byte `json:"-"`

	PackName  string `json:"packName"`
	GroupName string `json:"groupName"`
	Council   string `json:"council"`

	ContactLeaderFirstName   string      `json:"contactLeaderFirstName"`
	ContactLeaderLastName    string      `json:"contactLeaderLastName"`
	ContactLeaderPhoneNumber PhoneNumber `json:"contactLeaderPhoneNumber"`
	ContactLeaderAddress     Address     `json:"contactLeaderAddress"`

	ContactLeaderEmail string    `json:"contactLeaderEmail"`
	ValidatedOn        time.Time `json:"-"`
	ValidationToken    string    `json:"-"`

	EstimatedYouth   int `json:"estimatedYouth"`
	EstimatedLeaders int `json:"estimatedLeaders"`
}

func (gpr *GroupPreRegistration) Key() []byte {
	if gpr.SecurityKey == nil {
		var random [128]byte
		if _, err := rand.Read(random[:]); err != nil {
			panic(err)
		}
		gpr.SecurityKey = random[:]
	}
	return gpr.SecurityKey
}

func (gpr GroupPreRegistration) OrganicKey() string {
	return fmt.Sprintf("%s-%s-%s", strings.TrimSpace(gpr.Council), strings.TrimSpace(gpr.GroupName), strings.TrimSpace(gpr.PackName))
}

type PreRegDb interface {
	CreateRecord(rec *GroupPreRegistration) error
}

var (
	DBError             = errors.NewClass("Database Error")
	GroupAlreadyCreated = DBError.NewClass("Group already exists")
)

var (
	BOLT_GROUPBUCKET         = []byte("BUCKET_GROUP")
	BOLT_GROUPNAMEMAPBUCKET  = []byte("BUCKET_GROUPNAMEMAP")
	BOLT_GROUPEMAILMAPBUCKET = []byte("BUCKET_GROUPEMAILMAP")
)

type preRegDbBolt struct {
	db *bolt.DB
}

func NewPreRegBoltDb(db *bolt.DB) (PreRegDb, error) {
	prdb := &preRegDbBolt{
		db: db,
	}
	if err := prdb.init(); err != nil {
		return nil, err
	}
	return prdb, nil
}

func (d *preRegDbBolt) CreateRecord(in *GroupPreRegistration) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		gb := tx.Bucket(BOLT_GROUPBUCKET)
		gnmb := tx.Bucket(BOLT_GROUPNAMEMAPBUCKET)
		gemb := tx.Bucket(BOLT_GROUPEMAILMAPBUCKET)
		if gb.Get(in.Key()) != nil {
			log.Printf("Managed to get a duplicate security key somehow!")
			return GroupAlreadyCreated.New("Group with security key %v already exists", in.Key())
		} else if gnmb.Get([]byte(in.OrganicKey())) != nil {
			return GroupAlreadyCreated.New("Group with organic key %v already exists", in.OrganicKey())
		} else if gemb.Get([]byte(in.ContactLeaderEmail)) != nil {
			return GroupAlreadyCreated.New("Group with contact email %v already exists", in.ContactLeaderEmail)
		} else {
			data := &bytes.Buffer{}
			encoder := gob.NewEncoder(data)
			if err := encoder.Encode(in); err != nil {
				return err
			}
			if err := gb.Put(in.Key(), data.Bytes()); err != nil {
				return err
			}
			if err := gnmb.Put([]byte(in.OrganicKey()), in.Key()); err != nil {
				return err
			}
			if err := gemb.Put([]byte(in.ContactLeaderEmail), in.Key()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *preRegDbBolt) init() error {
	return d.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(BOLT_GROUPBUCKET); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(BOLT_GROUPNAMEMAPBUCKET); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(BOLT_GROUPEMAILMAPBUCKET); err != nil {
			return err
		}
		return nil
	})
}

type PreRegHandler struct {
	db PreRegDb
}

func (h *PreRegHandler) Create(w http.ResponseWriter, r *http.Request) {
	input := GroupPreRegistration{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		log.Printf("Got error while decoding json: %s", err)
		http.Error(w, "Invalid group json given", 400)
		return
	}

	if err := h.db.CreateRecord(&input); err != nil {
		log.Printf("Failed to insert record!  Error: %s", err)
		http.Error(w, "Failed to insert record", 500)
		return
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func registerGroupPreRegistrationHandler(r *mux.Router, db *bolt.DB) {
	prdb, err := NewPreRegBoltDb(db)
	if err != nil {
		log.Fatal(err)
	}
	preRegHandler := PreRegHandler{
		db: prdb,
	}

	r.HandleFunc("/prereg", preRegHandler.Create).Methods("POST")
}
