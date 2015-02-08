package main

import (
	"encoding/gob"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/google/google-api-go-client/plus/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type authStatusType int

const (
	authStatusLoggedIn authStatusType = 0
)

func init() {
	gob.Register(authStatusLoggedIn)
}

type AuthenticationHandler struct {
	store sessions.Store
}

func (a *AuthenticationHandler) VerifyGoogleToken(w http.ResponseWriter, r *http.Request) {
	var code string
	if codeB, err := ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Failed to read code!", http.StatusBadRequest)
		return
	} else {
		code = string(codeB)
	}
	conf := &oauth2.Config{
		ClientID:     authConfig.ClientID,
		ClientSecret: authConfig.ClientSecret,
		RedirectURL: "postmessage",
		Endpoint: google.Endpoint,
	}
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Print("Failed to exchange code, err ", err)
		http.Error(w, "Failed to talk to Google!", http.StatusInternalServerError)
		return
	}
	client := conf.Client(oauth2.NoContext, tok)

	plusService, err := plus.New(client)
	if err != nil {
		log.Panic("Shouldn't happen, err: ", err)
	}
	meGetter := plusService.People.Get("me")
	me, err := meGetter.Do()
	if err != nil {
		log.Print("Failed to get the user, err ", err)
		http.Error(w, "Failed to get user information!", http.StatusInternalServerError)
	}
	var primaryEmail string
	for _, emailInfo := range me.Emails {
		if emailInfo.Type == "account" {
			primaryEmail = emailInfo.Value
			break
		}
	}
	var validEmail bool
	for _, allowedEmail := range authConfig.AllowedEmails {
		if allowedEmail == primaryEmail {
			validEmail = true
			break
		}
	}
	if validEmail {
		sess, err := sessions.GetRegistry(r).Get(a.store, globalSessionName)
		if sess == nil {
			log.Panicf("Failed to get session, err %s", err)
		}
		sess.Values[authStatusLoggedIn] = true
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("true"))
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("false"))
	}
}

func (a *AuthenticationHandler) AdminFunc(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessions.GetRegistry(r).Get(a.store, globalSessionName)
		if sess == nil || sess.Values[authStatusLoggedIn] == nil || !sess.Values[authStatusLoggedIn].(bool) {
			log.Print("Failed to get session, err %s", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			h(w, r)
		}
	}
}

func NewAuthenticationHandler(r *mux.Router, store sessions.Store) *AuthenticationHandler {
	authHandler := &AuthenticationHandler{
		store: store,
	}

	r.HandleFunc("/authentication/googletoken", authHandler.VerifyGoogleToken).Methods("POST")

	return authHandler
}
