package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func adminSessionsPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle  string
		User       *patient
		Sessions   map[string]session
		ErrorMsg   string
		SuccessMsg string
	}{
		"Manage Sessions",
		thePatient,
		mapSessions,
		"",
		"",
	}

	if req.Method == http.MethodPost {
		// Get querystring values
		action := req.FormValue("action")
		sessionId := req.FormValue("sessionId")

		// delete single session
		if action == "delete" && sessionId != "" {
			if _, ok := mapSessions[sessionId]; ok {
				delete(mapSessions, sessionId)
				payload.SuccessMsg = "Session deleted!"
			} else {
				payload.ErrorMsg = ErrSessionNotFound.Error()
			}
		}

		// delete all sessions
		if action == "purge" {
			mapSessions = make(map[string]session)
		}

		if action != "" {
			// if user's session is gone, redirect away
			if _, isLoggedInCheck := isLoggedIn(req); !isLoggedInCheck {
				http.Redirect(res, req, pageLogin, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "adminSessions.gohtml", payload)
}

func adminAppointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle    string
		User         *patient
		Appointments []*appointment
		ErrorMsg     string
		SuccessMsg   string
	}{
		"Manage Appointments",
		thePatient,
		appointmentsSortedByTimeslot,
		"",
		"",
	}

	// Get notifications from session
	if notify, notifyErr := getNotification(req); notifyErr == nil {
		if notify != nil {
			if notify.Type == "Success" {
				payload.SuccessMsg = notify.Message
			} else if notify.Type == "Error" {
				payload.ErrorMsg = notify.Message
			}
			clearNotification(req)
		}
	}

	tpl.ExecuteTemplate(res, "adminAppointments.gohtml", payload)
}

func adminEditAppointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle          string
		User               *patient
		Appt               *appointment
		TimeslotsAvailable []int64
		ErrorMsg           string
	}{
		"Edit Appointment", thePatient, nil, nil, "",
	}

	// Get querystring values
	inputApptId := req.FormValue("apptId")
	action := req.FormValue("action")

	if action != "edit" && action != "cancel" {
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	apptId, err := strconv.ParseInt(inputApptId, 10, 64)

	if err != nil {
		setNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	// Check if appt id is valid
	theApptIndex := binarySearchApptID(apptId)

	if theApptIndex < 0 {
		setNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	payload.Appt = appointments[theApptIndex]

	// Cancel Appt
	if action == "cancel" {
		if req.Method == http.MethodPost {
			setNotification(req, "Appointment cancelled!", "Success")
			payload.Appt.cancelAppointment()
		}

		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	// Edit Appt
	if action == "edit" {

		payload.TimeslotsAvailable = getAvailableTimeslot(append(payload.Appt.Doctor.Appointments, payload.Appt.Patient.Appointments...))
		_, timeSlotErr := isThereTimeslot(payload.Appt.Patient, payload.Appt.Doctor)

		if timeSlotErr != nil {
			payload.ErrorMsg = timeSlotErr.Error()
			tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
			return
		}

		if req.Method == http.MethodPost {

			// Form submit values
			timeslot := req.FormValue("timeslot")

			if timeslot != "" {

				t, _ := strconv.ParseInt(timeslot, 10, 64)

				_, isApptTimeValidErr := isApptTimeValid(t) // Is time in the past check

				// Past time
				if isApptTimeValidErr != nil {
					payload.ErrorMsg = isApptTimeValidErr.Error()
					tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
					return
				}

				// Patient / Doctor time check
				if !payload.Appt.Patient.isFreeAt(t) || !payload.Appt.Doctor.isFreeAt(t) {
					payload.ErrorMsg = ErrDuplicateTimeslot.Error()
					tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
					return
				}

				payload.Appt.editAppointment(t, payload.Appt.Patient, payload.Appt.Doctor)
				setNotification(req, "Appointment updated!", "Success")
				http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
}

func adminUsersPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle  string
		User       *patient
		Patients   []*patient
		ErrorMsg   string
		SuccessMsg string
	}{
		"Manage Users",
		thePatient,
		patients,
		"",
		"",
	}

	if req.Method == http.MethodPost {
		// Get querystring values
		action := req.FormValue("action")
		userId := req.FormValue("userId")

		// delete single users
		if action == "delete" && userId != "" {
			theUser, err := getPatientByID(userId)

			if err == nil {
				theUser.deletePatient()
				payload.Patients = patients
				payload.SuccessMsg = "User deleted!"
			} else {
				payload.ErrorMsg = ErrPatientIDNotFound.Error()
			}
		}

		if action != "" {
			// if self is gone, bye bye
			if _, isLoggedInCheck := isLoggedIn(req); !isLoggedInCheck {
				http.Redirect(res, req, pageLogin, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "adminUsers.gohtml", payload)
}

func adminPaymentEnqueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		// Get querystring values
		apptId := req.FormValue("apptId")

		if apptId != "" {
			apptId, err := strconv.ParseInt(apptId, 10, 64)

			if err == nil {
				// Adding appt to global payment queue
				apptIdIndex := binarySearchApptID(apptId)

				if apptIdIndex < 0 {
					setNotification(req, "Error adding to payment queue! Appointment ID not found.", "Error")
					http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
					return
				}

				appt := appointments[apptIdIndex]
				createPayment(appt, 19.99)
				setNotification(req, "Appointment added to payment queue!", "Success")
				http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
				return
			}
		}
	}

	http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
}

func adminPaymentDequeuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		paymentQ.dequeue()
	}

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminPaymentDequeueToMissedQueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		paymentQ.dequeueToMissedPaymentQueue()
	}

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminPaymentDequeueToPaymentQueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		missedPaymentQ.dequeueToPaymentQueue()
	}

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
