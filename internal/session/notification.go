package session

import (
	"errors"
	"net/http"
)

type Notification struct {
	Message string
	Type    string // Types: "Success", "Error"
}

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

func GetNotification(req *http.Request) (*Notification, error) {
	myCookie, err := req.Cookie(CookieID)
	if err != nil {
		return nil, errors.New("unable to retrieve cookie")
	}

	return MapSessions[myCookie.Value].Notification, nil
}

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
