package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"bytes"
	"encoding/gob"
	"time"

	"github.com/boltdb/bolt"

	"github.com/spacemonkeygo/errors"
)

const keyLength = 24

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
	ValidatedOn        time.Time `json:"validatedOn"`
	ValidationToken    string    `json:"-"`

	EmailApprovalGivenAt time.Time `json:"emailApprovalGivenAt"`

	EmailConfirmationLastSendRequest time.Time `json:"-"`
	EmailConfirmationSent            bool      `json:"-"`
	EmailConfirmationSendErrors      int       `json:"-"`
	EmailConfirmationSendAttempts    int       `json:"-"`

	EstimatedYouth   int `json:"estimatedYouth"`
	EstimatedLeaders int `json:"estimatedLeaders"`
}

func (gpr GroupPreRegistration) Key() []byte {
	key, err := base64.URLEncoding.DecodeString(gpr.SecurityKey)
	if err != nil {
		panic("Invalid key")
	}
	if len(key) < keyLength {
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
	if !gpr.ValidatedOn.Equal(time.Time{}) {
		return RecordAlreadyPrepared.New("Email validation already given")
	}
	{
		var random [keyLength]byte
		if _, err := rand.Read(random[:]); err != nil {
			return err
		}
		gpr.SecurityKey = base64.URLEncoding.EncodeToString(random[:])
	}
	{
		var random [keyLength]byte
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

	NoteConfirmationEmailSent(rec *GroupPreRegistration) error
	VerifyEmail(email, token string) error
}

var (
	DBError               = errors.NewClass("Database Error")
	RecordAlreadyPrepared = DBError.NewClass("Group preregistration is already prepared")
	RecordDoesNotExist    = DBError.NewClass("Record does not exist")
	GroupAlreadyCreated   = DBError.NewClass("Group already exists")
	BadVerificationToken  = DBError.NewClass("Bad email verification token")
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

func insertUpdate(tx *bolt.Tx, update GroupPreRegistration, bucket *bolt.Bucket) error {
	data := &bytes.Buffer{}
	encoder := gob.NewEncoder(data)
	if err := encoder.Encode(update); err != nil {
		return err
	}
	if nextInt, err := bucket.NextSequence(); err != nil {
		return err
	} else {
		buf := new(bytes.Buffer)
		if err = binary.Write(buf, binary.BigEndian, nextInt); err != nil {
			return err
		} else if err = bucket.Put(buf.Bytes(), data.Bytes()); err != nil {
			return err
		}
	}
	return nil
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
			if bucket, err := gb.CreateBucketIfNotExists(in.Key()); err != nil {
				return err
			} else if err := insertUpdate(tx, *in, bucket); err != nil {
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

func fetchRecordWithNativeKey(tx *bolt.Tx, key []byte, origKey string) (rec *GroupPreRegistration, bucket *bolt.Bucket, err error) {
	gb := tx.Bucket(BOLT_GROUPBUCKET)
	bucket = gb.Bucket(key)
	if bucket == nil {
		return nil, nil, RecordDoesNotExist.New("Record for key %s does not exist", origKey)
	}
	_, data := bucket.Cursor().Last()
	if data == nil {
		return nil, nil, RecordDoesNotExist.New("Record for key %s does not exist", origKey)
	}
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var res GroupPreRegistration
	if err = decoder.Decode(&res); err != nil {
		return nil, nil, err
	} else {
		return &res, bucket, err
	}
}

func fetchRecordWithSecurityKey(tx *bolt.Tx, securityKey string) (rec *GroupPreRegistration, bucket *bolt.Bucket, err error) {
	key, err := base64.URLEncoding.DecodeString(securityKey)
	if err != nil {
		return nil, nil, err
	}
	return fetchRecordWithNativeKey(tx, key, securityKey)
}

func fetchRecordWithEmail(tx *bolt.Tx, email string) (rec *GroupPreRegistration, bucket *bolt.Bucket, err error) {
	gemb := tx.Bucket(BOLT_GROUPEMAILMAPBUCKET)
	key := gemb.Get([]byte(email))
	if key == nil {
		return nil, nil, RecordDoesNotExist.New("Record for key %s does not exist", email)
	}
	return fetchRecordWithNativeKey(tx, key, email)
}

func (d *preRegDbBolt) GetRecord(securityKey string) (rec *GroupPreRegistration, err error) {
	return rec, d.db.View(func(tx *bolt.Tx) error {
		res, _, err := fetchRecordWithSecurityKey(tx, securityKey)
		rec = res
		return err
	})
}

func (d *preRegDbBolt) NoteConfirmationEmailSent(gpr *GroupPreRegistration) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		rec, bucket, err := fetchRecordWithSecurityKey(tx, gpr.SecurityKey)
		if err != nil {
			return err
		}

		if rec.EmailConfirmationSent {
			return nil // Early return, avoid creating extra records.
		}

		rec.EmailConfirmationSent = true
		insertUpdate(tx, *rec, bucket)

		return nil
	})
	gpr.EmailConfirmationSent = true
	return err
}

func (d *preRegDbBolt) VerifyEmail(email, token string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		rec, bucket, err := fetchRecordWithEmail(tx, email)
		if err != nil {
			return BadVerificationToken.New("Failed to verify token")
		}

		if token != rec.ValidationToken {
			return BadVerificationToken.New("Failed to verify token")
		}

		if !rec.ValidatedOn.Equal(time.Time{}) {
			return nil // Early return, avoid creating extra records.
		}

		rec.ValidatedOn = time.Now()
		insertUpdate(tx, *rec, bucket)

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
	db                       PreRegDb
	confirmationEmailService *ConfirmationEmailService
	getHandler               *mux.Route
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
		// Finally, request an email confirmation.  Don't fail the create request if this fails.
		if err := h.confirmationEmailService.RequestEmailConfirmation(&input); err != nil {
			log.Printf("Failed to send initial email confirmation for key %s, error %s!", input.SecurityKey, err)
		}
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

func (h *PreRegHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email, ok := vars["email"]
	if !ok {
		http.Error(w, "No email given", 404)
		return
	}
	if tokenBytes, err := ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Failed to read token", http.StatusInternalServerError)
	} else {
		token := string(tokenBytes)
		if err = h.db.VerifyEmail(email, token); err != nil {
			http.Error(w, "Failed to verify token", http.StatusBadRequest)
		}
	}
	_ = email
}

func NewGroupPreRegistrationHandler(r *mux.Router, prdb PreRegDb, confirmationEmailService *ConfirmationEmailService) *PreRegHandler {
	preRegHandler := &PreRegHandler{
		db: prdb,
		confirmationEmailService: confirmationEmailService,
	}

	r.HandleFunc("/preregistration", preRegHandler.Create).Methods("POST")
	r.HandleFunc("/confirmpreregistration", preRegHandler.VerifyEmail).Queries("email", "{email:.*@.*}").Methods("PUT")
	preRegHandler.getHandler = r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}", preRegHandler.Get).Methods("GET")

	return preRegHandler
}
