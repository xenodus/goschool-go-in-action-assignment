package clinic

import (
	"errors"
)

var (
	// Appointments
	ErrInvalidTimeslot       = errors.New("invalid timeslots entered")
	ErrDoctorNoMoreTimeslot  = errors.New("doctor has no more timeslots available for today")
	ErrPatientNoMoreTimeslot = errors.New("patient has no more timeslots available for today")
	ErrNoMoreTimeslot        = errors.New("there are no other timeslots available with the chosen doctor")
	ErrDuplicateTimeslot     = errors.New("there's already an appointment scheduled for that timeslot")
	ErrTimeslotExpired       = errors.New("timeslot has already expired")
	ErrAppointmentIDNotFound = errors.New("appointment id not found")

	// Doc
	ErrDoctorIDNotFound = errors.New("doctor id not found")

	// Patient
	ErrPatientIDNotFound = errors.New("patient id not found")

	// Payment
	ErrEmptyPaymentQueue = errors.New("empty payment queue")
)
