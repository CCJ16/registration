package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"encoding/json"
	"io"
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
	SecurityKey string `json:"securityKey"`

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

	EmailApprovalGivenAt time.Time `json:"emailApprovalGivenAt,omit"`

	EmailConfirmationLastSent     time.Time `json:"-"`
	EmailConfirmationSendAttempts int       `json:"-"`

	EstimatedYouth   int `json:"estimatedYouth"`
	EstimatedLeaders int `json:"estimatedLeaders"`
}

func (gpr GroupPreRegistration) Key() []byte {
	key, err := base64.URLEncoding.DecodeString(gpr.SecurityKey)
	if err != nil {
		panic("Invalid key")
	}
	if len(key) < 129 {
		panic("Security was too short")
	}
	return key
}

func (gpr GroupPreRegistration) OrganicKey() string {
	return fmt.Sprintf("%s-%s-%s", strings.TrimSpace(gpr.Council), strings.TrimSpace(gpr.GroupName), strings.TrimSpace(gpr.PackName))
}

func (gpr *GroupPreRegistration) PrepareForInsert() error {
	if len(gpr.SecurityKey) != 0 {
		return RecordAlreadyPrepared.New("Security key has already been created, bailing out")
	}
	{
		var random [129]byte
		if _, err := rand.Read(random[:]); err != nil {
			return err
		}
		gpr.SecurityKey = base64.URLEncoding.EncodeToString(random[:])
	}
	{
		var random [129]byte
		if _, err := rand.Read(random[:]); err != nil {
			return err
		}
		gpr.ValidationToken = base64.URLEncoding.EncodeToString(random[:])
	}

	return nil
}

type PreRegDb interface {
	CreateRecord(rec *GroupPreRegistration) error
	GetRecord(securityKey string) (rec *GroupPreRegistration, err error)
}

var (
	DBError               = errors.NewClass("Database Error")
	RecordAlreadyPrepared = DBError.NewClass("Group preregistration is already prepared")
	RecordDoesNotExist    = DBError.NewClass("Record does not exist")
	GroupAlreadyCreated   = DBError.NewClass("Group already exists")
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
	if err := in.PrepareForInsert(); err != nil {
		return err
	}
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

func (d *preRegDbBolt) GetRecord(securityKey string) (rec *GroupPreRegistration, err error) {
	return rec, d.db.View(func(tx *bolt.Tx) error {
		gb := tx.Bucket(BOLT_GROUPBUCKET)
		key, err := base64.URLEncoding.DecodeString(securityKey)
		if err != nil {
			return err
		}
		data := gb.Get(key)
		if data == nil {
			return RecordDoesNotExist.New("Record for key %s does not exist", securityKey)
		}
		decoder := gob.NewDecoder(bytes.NewReader(data))
		var res GroupPreRegistration
		if err = decoder.Decode(&res); err != nil {
			return err
		} else {
			rec = &res
			return nil
		}
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
	db         PreRegDb
	getHandler *mux.Route
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
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(input); err != nil {
			http.Error(w, "Failed to insert record", 500)
			return
		}
		url, err := h.getHandler.URLPath("SecurityKey", input.SecurityKey)
		if err != nil {
			http.Error(w, "Failed to insert record", 500)
			return
		}
		w.Header()["Location"] = []string{url.Path}
		w.WriteHeader(http.StatusCreated)
		io.Copy(w, buf)
	}
}

func (h *PreRegHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	securityKey, ok := vars["SecurityKey"]
	if !ok {
		http.Error(w, "No key given", 404)
		return
	}
	rec, err := h.db.GetRecord(securityKey)
	if err != nil {
		http.Error(w, "Failed to get record", 500)
	} else {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(rec); err != nil {
			http.Error(w, "Failed to get record", 500)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.Copy(w, buf)
	}
}

func NewGroupPreRegistrationHandler(r *mux.Router, prdb PreRegDb) *PreRegHandler {
	preRegHandler := &PreRegHandler{
		db: prdb,
	}

	r.HandleFunc("/preregistration", preRegHandler.Create).Methods("POST")
	preRegHandler.getHandler = r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}", preRegHandler.Get).Methods("GET")

	return preRegHandler
}
