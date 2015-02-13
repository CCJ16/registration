package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/sessions"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestVerifySession(t *testing.T) {
	Convey("With initialized pieces", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:8080/api/authentication/isLoggedIn", nil)
		So(err, ShouldBeNil)
		store := sessions.NewCookieStore([]byte("A"))
		sess, err := sessions.GetRegistry(r).Get(store, globalSessionName)
		So(err, ShouldBeNil)
		So(sess, ShouldNotBeNil)
		aH := AuthenticationHandler{store}
		helper := func(output string) {
			w := httptest.NewRecorder()
			aH.VerifySession(w, r)
			w.Flush()
			So(w.HeaderMap.Get("Content-Type"), ShouldEqual, "text/plain")
			body, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, output)
		}
		Convey("Succeed when session is marked as true", func() {
			sess.Values[authStatusLoggedIn] = true
			helper("true")
		})
		Convey("Fail when session is marked as false", func() {
			sess.Values[authStatusLoggedIn] = false
			helper("false")
		})
		Convey("Fail when session is missing the value", func() {
			delete(sess.Values, authStatusLoggedIn)
			helper("false")
		})
	})
}
