package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type SummaryHandler struct {
	prdb PreRegDb
}

type PackSummaryOutput struct {
	YouthCount  int `json:"youthCount"`
	LeaderCount int `json:"leaderCount"`
}

func (sh *SummaryHandler) GetPack(w http.ResponseWriter, r *http.Request) {
	recs, err := sh.prdb.GetAll()
	if err != nil {
		httpError(w, err)
		return
	}

	output := PackSummaryOutput{}
	for i := 0; i < len(recs); i++ {
		output.YouthCount += recs[i].EstimatedYouth
		output.LeaderCount += recs[i].EstimatedLeaders
	}

	buf := &bytes.Buffer{}
	jsonEnc := json.NewEncoder(buf)
	err = jsonEnc.Encode(&output)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}
	w.Header()["Content-Type"] = []string{"application/json"}
	w.WriteHeader(http.StatusOK)
	io.Copy(w, buf)
}

func NewSummaryHandler(apiR *mux.Router, prdb PreRegDb) *SummaryHandler {
	sh := &SummaryHandler{
		prdb: prdb,
	}

	apiR.HandleFunc("/summary/pack", sh.GetPack).Methods("GET")

	return sh
}
