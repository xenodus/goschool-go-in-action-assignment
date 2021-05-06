package web

import (
	"assignment4/clinic"
	"errors"
	"strconv"
)

// Errors and accompanying messages to be output in logs or to users.
var (
	errInternalServerError = errors.New("internal server error")
	errStatusNotFound      = errors.New("page not found")

	// Session
	errSessionNotFound = errors.New("session not found")

	// Auth
	errAlreadyRegistered = errors.New("you are already registered")
	errInvalidNRIC       = errors.New("invalid NRIC")
	errAuthFailure       = errors.New("invalid NRIC / password combination")
	errMissingField      = errors.New("please enter all the fields")
	errPasswordLength    = errors.New("password length has to be >= " + strconv.Itoa(clinic.MinPasswordLength) + " characters")
)
