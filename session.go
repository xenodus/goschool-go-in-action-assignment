package main

import (
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
)

type session struct {
	Id           string
	LastModified int64
	LastVisited  *url.URL
	Notification *notification
}

// Globals
var cookieID string
var mapSessions = make(map[string]session)

func init() {
	// Randomizing the cookie name on each init
	cookieID = "AY_GOSCHOOL"
}

func deleteDuplicateSession(username string) {
	for k, v := range mapSessions {
		if v.Id == username {
			Info.Println("Stale session deleted successfully for:", username)
			delete(mapSessions, k)
			break
		}
	}
}

func createSession(res http.ResponseWriter, req *http.Request, username string) {

	deleteDuplicateSession(username)

	// Create Session + Cookie
	id, _ := uuid.NewV4()
	myCookie := &http.Cookie{
		Name:     cookieID,
		Value:    id.String(),
		Path:     pageIndex,
		HttpOnly: true,
		Secure:   true,
		SameSite: 3,
		Domain:   serverHost,
	}
	http.SetCookie(res, myCookie)
	mapSessions[myCookie.Value] = session{username, time.Now().Unix(), req.URL, nil}
}

func deleteSession(res http.ResponseWriter, req *http.Request) {
	myCookie, err := req.Cookie(cookieID)

	if err == nil {
		// Delete the Session
		delete(mapSessions, myCookie.Value)
		// Expire the Cookie
		expire := time.Now().Add(-7 * 24 * time.Hour)
		myCookie = &http.Cookie{
			Name:     cookieID,
			Value:    "",
			Path:     pageIndex,
			MaxAge:   -1,
			Expires:  expire,
			HttpOnly: true,
			Secure:   true,
			SameSite: 3,
		}
		http.SetCookie(res, myCookie)
	}
}
