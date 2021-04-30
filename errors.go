package main

import (
	"errors"
	"strconv"
)

var ErrInvalidTimeslot = errors.New("invalid timeslots entered")
var ErrDoctorNoMoreTimeslot = errors.New("doctor has no more timeslots available for today")
var ErrPatientNoMoreTimeslot = errors.New("patient has no more timeslots available for today")
var ErrNoMoreTimeslot = errors.New("there are no more timeslots available for today")
var ErrDuplicateTimeslot = errors.New("there's already an appointment scheduled for that timeslot")
var ErrTimeslotExpired = errors.New("timeslot has already expired")
var ErrAppointmentIDNotFound = errors.New("appointment id not found")
var ErrDoctorIDNotFound = errors.New("doctor id not found")
var ErrPatientIDNotFound = errors.New("patient id not found")
var ErrEmptyPaymentQueue = errors.New("empty payment queue")
var ErrSessionNotFound = errors.New("session not found")

var ErrInternalServerError = errors.New("internal server error")
var ErrStatusNotFound = errors.New("page not found")

// Auth
var ErrAlreadyRegistered = errors.New("you are already registered")
var ErrInvalidNRIC = errors.New("invalid NRIC")
var ErrAuthFailure = errors.New("invalid NRIC / password combination")
var ErrMissingField = errors.New("please enter all the fields")
var ErrPasswordLength = errors.New("password length has to be >= " + strconv.Itoa(minPasswordLength) + " characters")
