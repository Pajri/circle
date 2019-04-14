package utils

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

//these codes was taken from https://gowebexamples.com/sessions/
var store = sessions.NewCookieStore([]byte("session-key"))
var session *sessions.Session

//session keys
var SESSION_AUTH string = "session-auth"
var KEY_USERNAME string = "username"
var KEY_ISAUTH string = "is-auth"

func GetSession(r *http.Request, cookieName string) (*sessions.Session, error) {
	if session != nil {
		return session, nil
	}
	time.Now()
	store.Options = &sessions.Options{
		Path: "/",
	}

	var err error
	session, err = store.Get(r, cookieName)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func IsLoggedInSession(s *sessions.Session) bool {
	isAuthenticated := s.Values[KEY_ISAUTH]
	username := s.Values[KEY_USERNAME]

	if isAuthenticated == nil && username == nil {
		return false
	}

	if username.(string) != "" && isAuthenticated.(bool) {
		return true
	}
	return false
}

func GetUsernameFromSession(r *http.Request) (string, error) {
	var session *sessions.Session
	var err error
	session, err = GetSession(r, SESSION_AUTH)
	if err != nil {
		return "", err
	}
	username := session.Values[KEY_USERNAME].(string)
	return username, nil
}

func Login(s *sessions.Session, username string, w http.ResponseWriter, r *http.Request) error {
	s.Values[KEY_USERNAME] = username
	s.Values[KEY_ISAUTH] = true
	err := session.Save(r, w)
	return err
}

func Logout(w http.ResponseWriter, r *http.Request) error {
	if IsLoggedInSession(session) {
		session.Values[KEY_USERNAME] = ""
		session.Values[KEY_ISAUTH] = false
		err := session.Save(r, w)
		return err
	}
	return nil
}
