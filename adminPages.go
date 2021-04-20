package main

import (
	"fmt"
	"net/http"
	"strconv"
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
		mapSessions = make(map[string]session)
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
		Sessions  map[string]session
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

func adminEditAppointmentPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req) // the admin

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	// Get querystring values
	apptId := req.FormValue("apptId")
	action := req.FormValue("action")

	// Form submit values
	timeslot := req.FormValue("timeslot")

	var chosenDoctor *doctor = nil
	var theAppt *appointment = nil
	var timeslotsAvailable []int64
	var errorMsg = ""

	if action == "edit" || action == "cancel" {

		apptId, err := strconv.ParseInt(apptId, 10, 64)

		if err != nil {
			errorMsg = ErrAppointmentIDNotFound.Error()
		} else {
			// Check if appt id is valid
			theApptIndex := binarySearchApptID(appointments, 0, len(appointments)-1, apptId)

			if theApptIndex < 0 {
				errorMsg = ErrAppointmentIDNotFound.Error()
			} else {
				theAppt = appointments[theApptIndex]

				// Cancel Appt
				if action == "cancel" {
					cancelAppointment(apptId)
					http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
					return
				}

				// Edit Appt
				// Change thePatient to the actual patient since thePatient is Admin at the moment
				if action == "edit" {

					chosenDoctor = theAppt.Doctor
					timeslotsAvailable = getAvailableTimeslot(append(chosenDoctor.Appointments, theAppt.Patient.Appointments...))
					_, timeSlotErr := isThereTimeslot(theAppt.Patient, chosenDoctor)

					if timeSlotErr != nil {
						errorMsg = timeSlotErr.Error()
					}

					if timeslot != "" && chosenDoctor != nil {
						t, _ := strconv.ParseInt(timeslot, 10, 64)

						if isApptTimeValid, isApptTimeValidErr := isApptTimeValid(t); isApptTimeValid {
							theAppt.editAppointment(t, theAppt.Patient, chosenDoctor)
							http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
							return
						} else {
							errorMsg = isApptTimeValidErr.Error()
						}
					}
				}
			}
		}

		// Anonymous payload
		payload := struct {
			PageTitle          string
			User               *patient
			Doctors            []*doctor
			Appt               *appointment
			ChosenDoctor       *doctor
			TimeslotsAvailable []int64
			ErrorMsg           string
		}{
			"Edit Appointment",
			thePatient,
			doctors,
			theAppt,
			chosenDoctor,
			timeslotsAvailable,
			errorMsg,
		}

		tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
		return
	}

	http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
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
				createPayment(appt, 19.99)
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

func adminDebugPage(res http.ResponseWriter, req *http.Request) {

	fmt.Println(":::::::::::::: Debug Dump ::::::::::::::")
	fmt.Println("======================================")
	fmt.Println("Appointments:", len(appointments), appointments)
	fmt.Println("Appoinments sorted by time:", len(appointmentsSortedByTimeslot), appointmentsSortedByTimeslot)
	fmt.Println("Doctors:", len(doctors), doctors)
	fmt.Println("Doctors BST:", doctorsBST)
	fmt.Println("Patients:", len(patients), patients)
	fmt.Println("PaymentQueue:", paymentQ.Size, paymentQ)
	fmt.Println("MissedPaymentQueue:", missedPaymentQ.Size, missedPaymentQ)
	fmt.Println("Sessions:", len(mapSessions), mapSessions)
	fmt.Println("Admins:", len(admins), admins)

	fmt.Println(":::::::::::::: Appointments ::::::::::::::")
	fmt.Println("--- Id | Appt Time :::")

	for _, v := range appointments {
		fmt.Println(v.Id, time2HumanReadable(v.Time))
	}

	fmt.Println(":::::::::::::: Appointments By Time ::::::::::::::")
	fmt.Println("--- Id | Appt Time :::")

	for _, v := range appointmentsSortedByTimeslot {
		fmt.Println(v.Id, time2HumanReadable(v.Time))
	}

	fmt.Println(":::::::::::::: Doctors ::::::::::::::")
	fmt.Println("--- Id | # of Appts :::")

	for _, v := range doctors {
		fmt.Println(v.Id, len(v.Appointments))
	}

	fmt.Println(":::::::::::::: Patients ::::::::::::::")
	fmt.Println("--- Id | # of Appts :::")

	for _, v := range patients {
		fmt.Println(v.Id, len(v.Appointments))
	}

	http.Redirect(res, req, pageIndex, http.StatusSeeOther)
}
