package main

import (
	"encoding/json"
	"log"
	"net/http"

	"time"
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

type PreRegDb interface {
	CreateRecord(rec GroupPreRegistration) error
}

type PreRegDbBolt struct {
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

	if err := h.db.CreateRecord(input); err != nil {
		log.Printf("Failed to insert record!  Error: %s", err)
		http.Error(w, "Failed to insert record", 500)
		return
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func init() {
	preRegHandler := PreRegHandler{}

	r.HandleFunc("/prereg", preRegHandler.Create).Methods("POST")
}
