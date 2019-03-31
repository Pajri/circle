package utils

import (
	"net/http"
	"github.com/gorilla/sessions"
)

//these codes was taken from https://gowebexamples.com/sessions/
var store = sessions.NewCookieStore([]byte("session-key"))
var session *sessions.Session

//session keys
var SESSION_AUTH string = "session-auth"
var KEY_USERNAME string = "username"
var KEY_ISAUTH string= "is-auth"

func GetSession(r *http.Request, cookieName string) (*sessions.Session, error){
	if session != nil {
		return session, nil
	}

	store.Options = &sessions.Options{
		Path : "/",
	}
	
	var err error
	session, err = store.Get(r, cookieName)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func IsLoggedInSession(session *sessions.Session) bool{
	isAuthenticated := session.Values[KEY_ISAUTH]
	username := session.Values[KEY_USERNAME]

	if isAuthenticated == nil && username == nil {
		return false
	}

	if username.(string) != "" && isAuthenticated.(bool){
		return true
	}
	return false
}