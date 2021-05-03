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

	ErrCreateAppointment = errors.New("unable to create appointment")

	// Doc
	ErrDoctorIDNotFound = errors.New("doctor id not found")
	ErrCreateDoctor     = errors.New("unable to create doctor")

	// Patient
	ErrPatientIDNotFound = errors.New("patient id not found")
	ErrCreatePatient     = errors.New("unable to create patient")

	// Payment
	ErrEmptyPaymentQueue = errors.New("empty payment queue")

	// DB
	ErrDBConn = errors.New("unable to get db connection")
)
