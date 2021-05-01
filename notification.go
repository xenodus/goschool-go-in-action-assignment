package main

import (
	"errors"
	"net/http"
)

type notification struct {
	Message string
	Type    string // Types: "Success", "Error"
}

func setNotification(req *http.Request, notificationMsg, notificationType string) error {
	myCookie, err := req.Cookie(cookieID)
	if err != nil {
		return errors.New("unable to retrieve cookie")
	}

	session := mapSessions[myCookie.Value]
	notification := &notification{
		Message: notificationMsg,
		Type:    notificationType,
	}
	session.Notification = notification
	mapSessions[myCookie.Value] = session

	return nil
}

func getNotification(req *http.Request) (*notification, error) {
	myCookie, err := req.Cookie(cookieID)
	if err != nil {
		return nil, errors.New("unable to retrieve cookie")
	}

	return mapSessions[myCookie.Value].Notification, nil
}

func clearNotification(req *http.Request) error {
	myCookie, err := req.Cookie(cookieID)
	if err != nil {
		return errors.New("unable to retrieve cookie")
	}

	session := mapSessions[myCookie.Value]
	session.Notification = nil
	mapSessions[myCookie.Value] = session

	return nil
}
