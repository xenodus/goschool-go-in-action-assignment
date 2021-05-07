package session

import (
	"errors"
	"net/http"
)

type Notification struct {
	Message string
	Type    string // Types: "Success", "Error"
}

// SetNotification sets a notification message to user's session.
func SetNotification(req *http.Request, notificationMsg, notificationType string) error {
	myCookie, err := req.Cookie(CookieID)
	if err != nil {
		return errors.New("unable to retrieve cookie")
	}

	session := MapSessions[myCookie.Value]
	notification := &Notification{
		Message: notificationMsg,
		Type:    notificationType,
	}
	session.Notification = notification
	MapSessions[myCookie.Value] = session

	return nil
}

// GetNotification gets a notification message from user's session.
func GetNotification(req *http.Request) (*Notification, error) {
	myCookie, err := req.Cookie(CookieID)
	if err != nil {
		return nil, errors.New("unable to retrieve cookie")
	}

	return MapSessions[myCookie.Value].Notification, nil
}

// ClearNotification deletes a notification message from user's session.
func ClearNotification(req *http.Request) error {
	myCookie, err := req.Cookie(CookieID)
	if err != nil {
		return errors.New("unable to retrieve cookie")
	}

	session := MapSessions[myCookie.Value]
	session.Notification = nil
	MapSessions[myCookie.Value] = session

	return nil
}
