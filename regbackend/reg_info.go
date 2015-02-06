package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"bytes"
	"time"

	"github.com/CCJ16/registration/regbackend/boltorm"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
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

	InvoiceId uint64 `json:"invoiceId"`
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
	CreateInvoiceIfNotExists(rec *GroupPreRegistration) (inv *Invoice, err error)
}

var (
	DBError               = errors.NewClass("Database Error")
	DBGenericError        = DBError.NewClass("Database Error", errhttp.OverrideErrorBody("Please retry or contact site administrator"))
	RecordAlreadyPrepared = DBError.NewClass("Group preregistration is already prepared")
	RecordDoesNotExist    = DBError.NewClass("Record does not exist")
	GroupAlreadyCreated   = DBError.NewClass("Group already registered", errhttp.SetStatusCode(400))
	BadVerificationToken  = DBError.NewClass("Bad email verification token")
)

var (
	BOLT_GROUPBUCKET         = []byte("BUCKET_GROUP")
	BOLT_GROUPNAMEMAPBUCKET  = []byte("BUCKET_GROUPNAMEMAP")
	BOLT_GROUPEMAILMAPBUCKET = []byte("BUCKET_GROUPEMAILMAP")
)

type preRegDbBolt struct {
	db    boltorm.DB
	invDb InvoiceDb
}

func NewPreRegBoltDb(db boltorm.DB, invDb InvoiceDb) (PreRegDb, error) {
	prdb := &preRegDbBolt{
		db:    db,
		invDb: invDb,
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
	err := d.db.Update(func(tx boltorm.Tx) error {
		key := in.Key()
		if err := tx.Insert(BOLT_GROUPBUCKET, key, in); err != nil {
			return err
		} else if err := tx.AddIndex(BOLT_GROUPNAMEMAPBUCKET, []byte(in.OrganicKey()), key); err != nil {
			if boltorm.ErrKeyAlreadyExists.Contains(err) {
				return GroupAlreadyCreated.New("Group %s of %s, with pack name %s already exists", in.GroupName, in.Council, in.PackName)
			} else {
				return err
			}
		} else if err := tx.AddIndex(BOLT_GROUPEMAILMAPBUCKET, []byte(in.ContactLeaderEmail), key); err != nil {
			if boltorm.ErrKeyAlreadyExists.Contains(err) {
				return GroupAlreadyCreated.New("A previous group already registered with contact email address %s", in.ContactLeaderEmail)
			} else {
				return err
			}
		}
		return nil
	})
	if boltorm.ErrKeyAlreadyExists.Contains(err) {
		return GroupAlreadyCreated.New("Could not insert preregistration")
	} else {
		return err
	}
}

func (d *preRegDbBolt) GetRecord(securityKey string) (rec *GroupPreRegistration, err error) {
	return rec, d.db.View(func(tx boltorm.Tx) error {
		res := &GroupPreRegistration{}
		if key, err := base64.URLEncoding.DecodeString(securityKey); err != nil {
			return err
		} else if err = tx.Get(BOLT_GROUPBUCKET, key, res); err != nil {
			if boltorm.ErrKeyDoesNotExist.Contains(err) {
				return RecordDoesNotExist.New("Could not find preregistration")
			} else {
				return err
			}
		}
		rec = res
		return nil
	})
}

func (d *preRegDbBolt) NoteConfirmationEmailSent(gpr *GroupPreRegistration) error {
	err := d.db.Update(func(tx boltorm.Tx) error {
		rec := &GroupPreRegistration{}
		if err := tx.Get(BOLT_GROUPBUCKET, gpr.Key(), rec); err != nil {
			return err
		}

		if rec.EmailConfirmationSent {
			return nil // Early return, avoid creating extra records.
		}

		rec.EmailConfirmationSent = true
		return tx.Update(BOLT_GROUPBUCKET, rec.Key(), rec)
	})
	gpr.EmailConfirmationSent = true
	return err
}

func (d *preRegDbBolt) VerifyEmail(email, token string) error {
	return d.db.Update(func(tx boltorm.Tx) error {
		rec := &GroupPreRegistration{}
		if err := tx.GetByIndex(BOLT_GROUPEMAILMAPBUCKET, BOLT_GROUPBUCKET, []byte(email), rec); err != nil {
			return BadVerificationToken.New("Failed to verify token")
		}

		if token != rec.ValidationToken {
			return BadVerificationToken.New("Failed to verify token")
		}

		if !rec.ValidatedOn.Equal(time.Time{}) {
			return nil // Early return, avoid creating extra records.
		}

		rec.ValidatedOn = time.Now()
		return tx.Update(BOLT_GROUPBUCKET, rec.Key(), rec)
	})
}

func (d *preRegDbBolt) CreateInvoiceIfNotExists(gpr *GroupPreRegistration) (inv *Invoice, err error) {
	err = d.db.Update(func(tx boltorm.Tx) error {
		rec := &GroupPreRegistration{}
		if err := tx.Get(BOLT_GROUPBUCKET, gpr.Key(), rec); err != nil {
			return err
		}

		if rec.InvoiceId != 0 {
			if inv, err = d.invDb.GetInvoice(rec.InvoiceId, tx); err != nil {
				return err
			}
			gpr.InvoiceId = rec.InvoiceId
			return nil
		}

		var toLine string
		if rec.PackName == "" {
			toLine = fmt.Sprintf("%s of %s", rec.GroupName, rec.Council)
		} else {
			toLine = fmt.Sprintf("%s of %s (%s)", rec.GroupName, rec.Council, rec.PackName)
		}
		inv = &Invoice{
			To:        toLine,
			LineItems: []InvoiceItem{InvoiceItem{"Preregistration deposit", 25000, 1}},
		}
		if err := d.invDb.NewInvoice(inv, tx); err != nil {
			return err
		}
		rec.InvoiceId = inv.Id
		gpr.InvoiceId = inv.Id
		return tx.Update(BOLT_GROUPBUCKET, rec.Key(), rec)
	})
	return inv, err
}

func (d *preRegDbBolt) init() error {
	return d.db.Update(func(tx boltorm.Tx) error {
		if err := tx.CreateBucketIfNotExists(BOLT_GROUPBUCKET); err != nil {
			return err
		}
		if err := tx.CreateBucketIfNotExists(BOLT_GROUPNAMEMAPBUCKET); err != nil {
			return err
		}
		if err := tx.CreateBucketIfNotExists(BOLT_GROUPEMAILMAPBUCKET); err != nil {
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
		httpError(w, err)
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
		w.Header()["Content-Type"] = []string{"application/json"}
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
		w.Header()["Content-Type"] = []string{"application/json"}
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
}

func (h *PreRegHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	securityKey, ok := vars["SecurityKey"]
	if !ok {
		http.Error(w, "No key given", 404)
		return
	}
	preReg, err := h.db.GetRecord(securityKey)
	if err != nil {
		http.Error(w, "Failed to get record", 500)
	} else {
		inv, err := h.db.CreateInvoiceIfNotExists(preReg)
		if err != nil {
			httpError(w, err)
		}
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(inv); err != nil {
			http.Error(w, "Failed to get record", 500)
			return
		}
		w.Header()["Content-Type"] = []string{"application/json"}
		w.WriteHeader(http.StatusOK)
		io.Copy(w, buf)
	}
}

func NewGroupPreRegistrationHandler(r *mux.Router, prdb PreRegDb, confirmationEmailService *ConfirmationEmailService) *PreRegHandler {
	preRegHandler := &PreRegHandler{
		db: prdb,
		confirmationEmailService: confirmationEmailService,
	}

	r.HandleFunc("/preregistration", preRegHandler.Create).Methods("POST")
	r.HandleFunc("/confirmpreregistration", preRegHandler.VerifyEmail).Queries("email", "{email:.*@.*}").Methods("PUT")
	preRegHandler.getHandler = r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}", preRegHandler.Get).Methods("GET")
	r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}/invoice", preRegHandler.GetInvoice).Methods("GET")

	return preRegHandler
}
