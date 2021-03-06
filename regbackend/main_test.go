package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testHttpHandler struct{}

func (testHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(299)
}

func verifyCookie(c *http.Cookie, name string, httpOnly bool) func() {
	return func() {
		Convey("Name", func() {
			So(c.Name, ShouldEqual, name)
		})
		Convey("httpOnly", func() {
			So(c.HttpOnly, ShouldEqual, httpOnly)
		})
		Convey("secure", func() {
			So(c.Secure, ShouldEqual, true)
		})
		Convey("path", func() {
			So(c.Path, ShouldEqual, "/")
		})
		Convey("expires", func() {
			So(c.Expires, ShouldHappenWithin, 1*time.Second, time.Now().Add(24*30*time.Hour))
		})
		Convey("max-age", func() {
			So(c.MaxAge, ShouldEqual, 24*30*60*60)
		})
	}
}

func TestXsrfSetting(t *testing.T) {
	Convey("With a successful setXsrfToken call (assuming production mode)", t, func() {
		r, err := http.NewRequest("GET", "http://localhost/", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		store := sessions.NewCookieStore()
		(&xsrfTokenCreator{store: store, config: &configType{}, Handler: testHttpHandler{}}).ServeHTTP(w, r)
		w.Flush()
		So(w.Code, ShouldEqual, 299)
		Reset(func() {
			context.Clear(r)
		})
		Convey("And the cookie is set after flushing", func() {
			resp := &http.Response{
				Header: w.HeaderMap,
			}
			cookies := resp.Cookies()
			Convey("With only my cookie set", func() {
				So(len(cookies), ShouldEqual, 1)
				var jsCookie *http.Cookie
				for _, cookie := range cookies {
					switch cookie.Name {
					case "XSRF-TOKEN":
						jsCookie = cookie
					}
				}
				So(jsCookie, ShouldNotBeNil)
				Convey("And the session variable created", func() {
					sess, err := sessions.GetRegistry(r).Get(store, globalSessionName)
					So(err, ShouldBeNil)
					sessionValue := sess.Values[xsrfSessionToken].(string)
					Convey("With matching values", func() {
						So(jsCookie.Value, ShouldEqual, sessionValue)
					})
					Convey("With the js cookie setup right", verifyCookie(jsCookie, "XSRF-TOKEN", false))
				})
			})
		})
	})
}

func TestXsrfVerifications(t *testing.T) {
	Convey("With setup", t, func() {
		var sessionValue interface{}
		header := make(http.Header)
		verifyResult := func(succeed bool) func() {
			return func() {
				r := &http.Request{
					Header: header,
				}
				Reset(func() {
					context.Clear(r)
				})
				store := sessions.NewCookieStore()
				session, err := store.Get(r, globalSessionName)
				So(err, ShouldBeNil)
				session.Values[xsrfSessionToken] = sessionValue
				w := httptest.NewRecorder()
				xsrfHandler := &xsrfVerifierHandler{&xsrfTokenCreator{nil, &configType{}, store}, testHttpHandler{}}
				xsrfHandler.ServeHTTP(w, r)
				w.Flush()
				if succeed {
					So(w.Code, ShouldEqual, 299)
				} else {
					So(w.Code, ShouldEqual, 400)
				}
			}
		}
		Convey("With matching header/cookie values with good names", func() {
			sessionValue = "Token"
			header.Add("X-Xsrf-Token", "Token")
			Convey("Succeeds", verifyResult(true))
		})
		Convey("With matching header/cookie values with a bad header name", func() {
			sessionValue = "Token"
			header.Add("X-Non-Token-Name", "Token")
			Convey("Fails", verifyResult(false))
		})
		Convey("With missing session data", func() {
			header.Add("X-Xsrf-Token", "Token")
			Convey("Fails", verifyResult(false))
		})
		Convey("With mismatching header/cookie values with good names", func() {
			sessionValue = "Token"
			header.Add("X-Non-Token-Name", "OtherToken")
			Convey("Fails", verifyResult(false))
		})
		Convey("With missing header", func() {
			sessionValue = "Token"
			Convey("Fails", verifyResult(false))
		})
	})
}

func TestStringSliceConfig(t *testing.T) {
	Convey("With a valid stringSliceConfig object", t, func() {
		var ssc stringSliceConfig
		Convey("Parsing a string with comma separated values returns the string split", func() {
			str := "test,string,values,nums223,email@email.com"
			ssc.Set(str)
			So(ssc, ShouldResemble, stringSliceConfig{"test", "string", "values", "nums223", "email@email.com"})
			Convey("And stringifying that should return the original string surrounded by quotes", func() {
				So(ssc.String(), ShouldEqual, "\""+str+"\"")
			})
		})
	})
}

func TestConfigHandler(t *testing.T) {
	Convey("With setup", t, func() {
		verifyResult := func(wait, open bool) func() {
			return func() {
				r, err := http.NewRequest("GET", "http://localhost:8080/config", nil)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()
				c := &configType{}
				c.General.EnableWaitingList = wait
				c.General.EnableGroupReg = open
				handler := configHandler{c}
				handler.ServeHTTP(w, r)
				w.Flush()
				So(w.Code, ShouldEqual, 200)
				m := make(map[string]bool)
				json.Unmarshal(w.Body.Bytes(), &m)
				v, ok := m["registrationOpen"]
				So(ok, ShouldBeTrue)
				So(v, ShouldEqual, open)
				v, ok = m["registrationOnWaitingList"]
				So(ok, ShouldBeTrue)
				So(v, ShouldEqual, wait)
			}
		}
		Convey("with disabled waiting list and disabled registration", func() {
			Convey("Succeeds", verifyResult(false, false))
		})
		Convey("with enabled waiting list and disabled registration", func() {
			Convey("Succeeds", verifyResult(true, false))
		})
		Convey("with disabled waiting list and enabled registration", func() {
			Convey("Succeeds", verifyResult(false, true))
		})
		Convey("with enabled waiting list and enabled registration", func() {
			Convey("Succeeds", verifyResult(true, true))
		})
	})
}

func TestDisableCacheHandler(t *testing.T) {
	Convey("With setup", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:8080/config", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		handler := disableCacheHandler{&configHandler{&configType{}}}
		handler.ServeHTTP(w, r)
		w.Flush()
		So(w.Code, ShouldEqual, 200)
		Convey("The header is present", func() {
			So(w.HeaderMap.Get("Cache-Control"), ShouldEqual, "no-cache, no-store")
		})
		Convey("And config handler was called and returned its data", func() {
			So(w.Body.Len(), ShouldNotEqual, 0)
		})
	})
}

func init() {
	// Disable log output for now, to clean up output.
	log.SetOutput(ioutil.Discard)
}
