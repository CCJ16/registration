package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"encoding/binary"
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

	IsOnWaitingList bool `json:"isOnWaitingList"`

	InvoiceID uint64 `json:"invoiceId"`
}

type GroupPreRegistrationInWaitingList struct {
	*GroupPreRegistration
	WaitingListPos int `json:"waitingListPos"`
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
	GetAll() (recs []*GroupPreRegistration, err error)
	GetWaitingList() (recs []*GroupPreRegistrationInWaitingList, err error)

	NoteConfirmationEmailSent(rec *GroupPreRegistration) error
	VerifyEmail(email, token string) error
	CreateInvoiceIfNotExists(rec *GroupPreRegistration) (inv *Invoice, err error)
	Promote(securityKey string) error
}

var (
	DBError                = errors.NewClass("Database Error")
	DBGenericError         = DBError.NewClass("Database Error", errhttp.OverrideErrorBody("Please retry or contact site administrator"))
	RecordAlreadyPrepared  = DBError.NewClass("Group preregistration is already prepared")
	RecordDoesNotExist     = DBError.NewClass("Record does not exist")
	GroupAlreadyCreated    = DBError.NewClass("Group already registered", errhttp.SetStatusCode(400))
	BadVerificationToken   = DBError.NewClass("Bad email verification token")
	NoInvoiceOnWaitingList = DBError.NewClass("No payments are collected on the waiting list", errhttp.SetStatusCode(400))
	NotOnWaitingList       = DBError.NewClass("Record is already not on the waiting list", errhttp.SetStatusCode(400))
)

var (
	BOLT_GROUPBUCKET             = []byte("BUCKET_GROUP")
	BOLT_GROUPNAMEMAPBUCKET      = []byte("BUCKET_GROUPNAMEMAP")
	BOLT_GROUPEMAILMAPBUCKET     = []byte("BUCKET_GROUPEMAILMAP")
	BOLT_GROUPEWAITINGLISTBUCKET = []byte("BUCKET_GROUPEWAITINGLIST")
)

type preRegDbBolt struct {
	db     boltorm.DB
	config *configType
	invDb  InvoiceDb
}

func NewPreRegBoltDb(db boltorm.DB, config *configType, invDb InvoiceDb) (PreRegDb, error) {
	prdb := &preRegDbBolt{
		db:     db,
		config: config,
		invDb:  invDb,
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
	in.IsOnWaitingList = d.config.General.EnableWaitingList

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

		if in.IsOnWaitingList {
			waitingListPos, err := tx.NextSequenceForBucket(BOLT_INVOICEBUCKET)
			if err != nil {
				return err
			}
			var waitingListPosBytes [8]byte
			binary.BigEndian.PutUint64(waitingListPosBytes[:], waitingListPos)
			if err := tx.AddIndex(BOLT_GROUPEWAITINGLISTBUCKET, waitingListPosBytes[:], key); err != nil {
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

func (d *preRegDbBolt) getRecord(tx boltorm.Tx, securityKey string) (rec *GroupPreRegistration, err error) {
	rec = &GroupPreRegistration{}
	if key, err := base64.URLEncoding.DecodeString(securityKey); err != nil {
		return nil, err
	} else if err = tx.Get(BOLT_GROUPBUCKET, key, rec); err != nil {
		if boltorm.ErrKeyDoesNotExist.Contains(err) {
			return nil, RecordDoesNotExist.New("Could not find preregistration")
		} else {
			return nil, err
		}
	}
	return rec, nil
}

func (d *preRegDbBolt) GetRecord(securityKey string) (rec *GroupPreRegistration, err error) {
	return rec, d.db.View(func(tx boltorm.Tx) error {
		rec, err = d.getRecord(tx, securityKey)
		return err
	})
}

func (d *preRegDbBolt) GetAll() (recs []*GroupPreRegistration, err error) {
	return recs, d.db.View(func(tx boltorm.Tx) error {
		if res, err := tx.GetAll(BOLT_GROUPBUCKET, &GroupPreRegistration{}); err != nil {
			return err
		} else {
			recs = res.([]*GroupPreRegistration)
		}
		return nil
	})
}

func (d *preRegDbBolt) GetWaitingList() (recs []*GroupPreRegistrationInWaitingList, err error) {
	return recs, d.db.View(func(tx boltorm.Tx) error {
		if res, err := tx.GetAllByIndex(BOLT_GROUPEWAITINGLISTBUCKET, BOLT_GROUPBUCKET, &GroupPreRegistration{}); err != nil {
			return err
		} else {
			rawRecs := res.([]*GroupPreRegistration)
			for i, rec := range rawRecs {
				recs = append(recs, &GroupPreRegistrationInWaitingList{
					rec,
					i + 1,
				})
			}
		}
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

		if rec.IsOnWaitingList {
			return NoInvoiceOnWaitingList.New("You are currently on the waiting list")
		}

		if rec.InvoiceID != 0 {
			if inv, err = d.invDb.GetInvoice(rec.InvoiceID, tx); err != nil {
				return err
			}
			gpr.InvoiceID = rec.InvoiceID
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
			LineItems: []InvoiceItem{{"Pre-registration deposit", 25000, 1}},
		}
		if err := d.invDb.NewInvoice(inv, tx); err != nil {
			return err
		}
		rec.InvoiceID = inv.ID
		gpr.InvoiceID = inv.ID
		return tx.Update(BOLT_GROUPBUCKET, rec.Key(), rec)
	})
	return inv, err
}

func (d *preRegDbBolt) Promote(securityKey string) error {
	return d.db.Update(func(tx boltorm.Tx) error {
		rec, err := d.getRecord(tx, securityKey)
		if err != nil {
			return err
		}

		if !rec.IsOnWaitingList {
			return NotOnWaitingList.New(securityKey + " is not on the waiting list!")
		}

		// Ok, this record is ready to move.  Change its flag and remove from the index.
		rec.IsOnWaitingList = false
		if err := tx.RemoveKeyFromIndex(BOLT_GROUPEWAITINGLISTBUCKET, rec.Key()); err != nil {
			return err
		}
		if err := tx.Update(BOLT_GROUPBUCKET, rec.Key(), rec); err != nil {
			return err
		}
		return nil
	})
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
		if err := tx.CreateBucketIfNotExists(BOLT_GROUPEWAITINGLISTBUCKET); err != nil {
			return err
		}
		return nil
	})
}

type PreRegHandler struct {
	db                       PreRegDb
	config                   *configType
	confirmationEmailService *ConfirmationEmailService
	getHandler               *mux.Route
}

func (h *PreRegHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.config.General.EnableGroupReg {
		http.Error(w, "Group registrations are closed", http.StatusForbidden)
		return
	}

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

func (h *PreRegHandler) GetList(w http.ResponseWriter, r *http.Request) {
	const (
		allRegs        = "all"
		registeredRegs = "registered"
		waitingRegs    = "waiting"
	)

	selectValue := r.URL.Query().Get("select")
	// Only all/registered/waiting are valid selections, if that isn't passed assume all.
	if selectValue != allRegs && selectValue != registeredRegs && selectValue != waitingRegs {
		selectValue = allRegs
	}

	var output interface{}
	if selectValue == allRegs || selectValue == registeredRegs {
		recs, err := h.db.GetAll()
		if err != nil {
			http.Error(w, "Failed to get records", 500)
			return
		} else {
			if selectValue == "all" {
				output = recs
			} else {
				filteredRecs := []*GroupPreRegistration{}
				for _, rec := range recs {
					if !rec.IsOnWaitingList {
						filteredRecs = append(filteredRecs, rec)
					}
				}
				output = filteredRecs
			}
		}
	} else {
		recs, err := h.db.GetWaitingList()
		if err != nil {
			http.Error(w, "Failed to get records", 500)
			return
		} else {
			output = recs
		}
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(output); err != nil {
		http.Error(w, "Failed to get records", 500)
		return
	}
	w.Header()["Content-Type"] = []string{"application/json"}
	w.WriteHeader(http.StatusOK)
	io.Copy(w, buf)
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
			return
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

func (h *PreRegHandler) PromoteToRegistration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	securityKey, ok := vars["SecurityKey"]
	if !ok {
		http.Error(w, "No key given", 404)
		return
	}
	err := h.db.Promote(securityKey)
	if err != nil {
		httpError(w, err)
		return
	}
}

func NewGroupPreRegistrationHandler(r *mux.Router, config *configType, prdb PreRegDb, authHandler *AuthenticationHandler, confirmationEmailService *ConfirmationEmailService) *PreRegHandler {
	preRegHandler := &PreRegHandler{
		db:     prdb,
		config: config,
		confirmationEmailService: confirmationEmailService,
	}

	r.HandleFunc("/preregistration", preRegHandler.Create).Methods("POST")
	r.HandleFunc("/confirmpreregistration", preRegHandler.VerifyEmail).Queries("email", "{email:.*@.*}").Methods("PUT")
	preRegHandler.getHandler = r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}", preRegHandler.Get).Methods("GET")
	r.HandleFunc("/preregistration", authHandler.AdminFunc(preRegHandler.GetList)).Methods("Get")
	r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}/invoice", preRegHandler.GetInvoice).Methods("GET")
	r.HandleFunc("/preregistration/{SecurityKey:[a-zA-Z0-9-_]+}/promote", authHandler.AdminFunc(preRegHandler.PromoteToRegistration)).Methods("POST")

	return preRegHandler
}
