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
}

// Globals
var cookieID string
var mapSessions = make(map[string]session)

func init() {
	// Randomizing the cookie name on each init
	cookieID = getRandomCookiePrefix()
}

func createSession(res http.ResponseWriter, req *http.Request, username string) {
	// Create Session + Cookie
	id, _ := uuid.NewV4()
	myCookie := &http.Cookie{
		Name:     cookieID,
		Value:    id.String(),
		Path:     pageIndex,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(res, myCookie)
	mapSessions[myCookie.Value] = session{username, time.Now().Unix(), req.URL}
}

func deleteSession(res http.ResponseWriter, req *http.Request) {
	myCookie, _ := req.Cookie(cookieID)
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
	}
	http.SetCookie(res, myCookie)
}

func getRandomCookiePrefix() string {
	randomUUIDByte, _ := uuid.NewV4()
	return "CO-" + randomUUIDByte.String()
}
