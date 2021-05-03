package clinic

import (
	"errors"
)

var ErrInvalidTimeslot = errors.New("invalid timeslots entered")
var ErrDoctorNoMoreTimeslot = errors.New("doctor has no more timeslots available for today")
var ErrPatientNoMoreTimeslot = errors.New("patient has no more timeslots available for today")
var ErrNoMoreTimeslot = errors.New("there are no other timeslots available with the chosen doctor")
var ErrDuplicateTimeslot = errors.New("there's already an appointment scheduled for that timeslot")
var ErrTimeslotExpired = errors.New("timeslot has already expired")
var ErrAppointmentIDNotFound = errors.New("appointment id not found")
var ErrDoctorIDNotFound = errors.New("doctor id not found")
var ErrPatientIDNotFound = errors.New("patient id not found")
var ErrEmptyPaymentQueue = errors.New("empty payment queue")
var ErrSessionNotFound = errors.New("session not found")
