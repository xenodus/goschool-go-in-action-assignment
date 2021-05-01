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

func createSession(res http.ResponseWriter, req *http.Request, username string) {
	// Create session
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
	// Delete the session
	delete(mapSessions, myCookie.Value)
	// Remove the cookie
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
