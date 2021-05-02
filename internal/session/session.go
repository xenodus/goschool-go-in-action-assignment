package session

import (
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Session struct {
	Id           string
	LastModified int64
	LastVisited  *url.URL
	Notification *Notification
}

// Globals
var CookieID string
var MapSessions = make(map[string]Session)

func init() {
	CookieID = "AY_GOSCHOOL"
}

func deleteDuplicateSession(username string) {
	for k, v := range MapSessions {
		if v.Id == username {
			//Info.Println("Stale session deleted successfully for:", username)
			delete(MapSessions, k)
			break
		}
	}
}

func CreateSession(res http.ResponseWriter, req *http.Request, username, serverHost string) {

	deleteDuplicateSession(username)

	// Create Session + Cookie
	id, _ := uuid.NewV4()
	myCookie := &http.Cookie{
		Name:     CookieID,
		Value:    id.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: 3,
		Domain:   serverHost,
	}
	http.SetCookie(res, myCookie)
	MapSessions[myCookie.Value] = Session{username, time.Now().Unix(), req.URL, nil}
}

func DeleteSession(res http.ResponseWriter, req *http.Request) {
	myCookie, err := req.Cookie(CookieID)

	if err == nil {
		// Delete the Session
		delete(MapSessions, myCookie.Value)
		// Expire the Cookie
		expire := time.Now().Add(-7 * 24 * time.Hour)
		myCookie = &http.Cookie{
			Name:     CookieID,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  expire,
			HttpOnly: true,
			Secure:   true,
			SameSite: 3,
		}
		http.SetCookie(res, myCookie)
	}
}
