package session

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Session struct {
	Id           string
	LastModified int64
	LastVisited  *url.URL
	Notification *Notification
}

// CookieID is the name of the client side cookie.
var CookieID string

// MapSessions stores all the user session(s) of the app.
var MapSessions = make(map[string]Session)

func init() {
	CookieID = "AY_GOSCHOOL"
}

func deleteDuplicateSession(username string) {

	file, err := os.OpenFile("./logs/out.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	defer file.Close()

	info := log.New(io.MultiWriter(os.Stdout, file), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	for k, v := range MapSessions {
		if v.Id == username {
			info.Println("Stale session deleted successfully for:", username)
			delete(MapSessions, k)
			break
		}
	}
}

// Creates client side cookie, create and add session to global MapSessions; Also, purge any duplicate sessions by a user.
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

// Expires a user's client side cookie and remove session from global MapSessions.
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
