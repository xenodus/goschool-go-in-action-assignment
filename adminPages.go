package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

func adminSessionsPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	var errorMsg = ""

	// Get querystring values
	action := req.FormValue("action")
	sessionId := req.FormValue("sessionId")

	// delete single session
	if action == "delete" && sessionId != "" {
		if _, ok := mapSessions[sessionId]; ok {
			delete(mapSessions, sessionId)
		} else {
			errorMsg = ErrSessionNotFound.Error()
		}
	}

	// delete all sessions
	if action == "purge" {
		mapSessions = make(map[string]string)
	}

	if action != "" {
		// if user's session is gone, redirect away
		if !isLoggedIn(req) {
			http.Redirect(res, req, pageLogin, http.StatusSeeOther)
			return
		}
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
		Sessions  map[string]string
		ErrorMsg  string
	}{
		"Manage Sessions",
		thePatient,
		mapSessions,
		errorMsg,
	}

	tpl.ExecuteTemplate(res, "adminSessions.gohtml", payload)
}

func adminAppointmentPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	var errorMsg = ""

	// Get querystring values
	err := req.FormValue("error")

	if err == "ErrAppointmentIDNotFound" {
		errorMsg = ErrAppointmentIDNotFound.Error()
	}

	// Anonymous payload
	payload := struct {
		PageTitle    string
		User         *patient
		Appointments []*appointment
		ErrorMsg     string
	}{
		"Manage Appointments",
		thePatient,
		appointmentsSortedByTimeslot,
		errorMsg,
	}

	tpl.ExecuteTemplate(res, "adminAppointments.gohtml", payload)
}

func adminUsersPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	var errorMsg = ""

	// Get querystring values
	action := req.FormValue("action")
	userId := req.FormValue("userId")

	// delete single users
	if action == "delete" && userId != "" {
		theUser, err := getPatientByID(userId)

		if err == nil {
			theUser.delete()
		} else {
			errorMsg = ErrPatientIDNotFound.Error()
		}
	}

	if action != "" {
		// if self is gone, bye bye
		if !isLoggedIn(req) {
			http.Redirect(res, req, pageLogin, http.StatusSeeOther)
			return
		}
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
		Patients  []*patient
		ErrorMsg  string
	}{
		"Manage Users",
		thePatient,
		patients,
		errorMsg,
	}

	tpl.ExecuteTemplate(res, "adminUsers.gohtml", payload)
}

func adminPaymentEnqueuePage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	// Get querystring values
	apptId := req.FormValue("apptId")

	if apptId != "" {
		apptId, err := strconv.ParseInt(apptId, 10, 64)

		if err == nil {
			// Adding appt to global payment queue
			apptIdIndex := binarySearchApptID(appointments, 0, len(appointments)-1, apptId)

			if apptIdIndex >= 0 {
				appt := appointments[apptIdIndex]
				atomic.AddInt64(&paymentCounter, 1)
				pmy := payment{paymentCounter, appt, 19.99} // yup... flat rate
				paymentQ.enqueue(&pmy)
				cancelAppointment(apptId)
				http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
				return
			}
		}
	}

	http.Redirect(res, req, pageAdminAllAppointments+"?error=ErrAppointmentIDNotFound", http.StatusSeeOther)
}

func adminPaymentDequeuePage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	paymentQ.dequeue()

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminPaymentDequeueToMissedQueuePage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	paymentQ.dequeueToMissedPaymentQueue()

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminPaymentDequeueToPaymentQueuePage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	missedPaymentQ.dequeueToPaymentQueue()

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}
