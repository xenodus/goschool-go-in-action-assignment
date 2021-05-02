package main

import (
	"fmt"
	"net/http"
	"strconv"

	"./internal/session"
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
		Sessions   map[string]session.Session
		ErrorMsg   string
		SuccessMsg string
	}{
		"Manage Sessions",
		thePatient,
		session.MapSessions,
		"",
		"",
	}

	if req.Method == http.MethodPost {
		// Get querystring values
		action := req.FormValue("action")
		sessionId := req.FormValue("sessionId")

		// delete single session
		if action == "delete" && sessionId != "" {
			if _, ok := session.MapSessions[sessionId]; ok {
				delete(session.MapSessions, sessionId)
				payload.SuccessMsg = "Session deleted!"
				Info.Println(req.RemoteAddr, " [Admin] Session deleted successfully. By:", thePatient.Id)
			} else {
				payload.ErrorMsg = ErrSessionNotFound.Error()
				Error.Println(req.RemoteAddr, "[Admin] Session delete failure:", payload.ErrorMsg, "By:", thePatient.Id)
			}
		}

		// delete all sessions
		if action == "purge" {
			session.MapSessions = make(map[string]session.Session)
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
	if notify, notifyErr := session.GetNotification(req); notifyErr == nil {
		if notify != nil {
			if notify.Type == "Success" {
				payload.SuccessMsg = notify.Message
			} else if notify.Type == "Error" {
				payload.ErrorMsg = notify.Message
			}
			session.ClearNotification(req)
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
		Error.Println(req.RemoteAddr, " [Admin] Appointment update failure: invalid action type. By:", thePatient.Id)
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	apptId, err := strconv.ParseInt(inputApptId, 10, 64)

	if err != nil {
		session.SetNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		Error.Println(req.RemoteAddr, " [Admin] Appointment update failure: invalid appt id. Unable to parse. By:", thePatient.Id)
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	// Check if appt id is valid
	theApptIndex := binarySearchApptID(apptId)

	if theApptIndex < 0 {
		session.SetNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		Error.Println(req.RemoteAddr, " [Admin] Appointment update failure: ", ErrAppointmentIDNotFound.Error())
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	payload.Appt = appointments[theApptIndex]

	// Cancel Appt
	if action == "cancel" {
		if req.Method == http.MethodPost {
			session.SetNotification(req, "Appointment cancelled!", "Success")
			Info.Println(req.RemoteAddr, " [Admin] Appointment cancelled successfully:", payload.Appt.Id, "By:", thePatient.Id)
			payload.Appt.cancelAppointment()
		} else {
			Error.Println(req.RemoteAddr, " [Admin] Appointment cancellation failure: GET REQUEST. By:", thePatient.Id)
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
			Error.Println(req.RemoteAddr, " [Admin] Appointment update failure:", payload.ErrorMsg, "By:", thePatient.Id)
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
					Error.Println(req.RemoteAddr, " [Admin] Appointment update failure:", payload.ErrorMsg, "By:", thePatient.Id)
					tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
					return
				}

				// Patient / Doctor time check
				if !payload.Appt.Patient.isFreeAt(t) || !payload.Appt.Doctor.isFreeAt(t) {
					payload.ErrorMsg = ErrDuplicateTimeslot.Error()
					Error.Println(req.RemoteAddr, " [Admin] Appointment update failure:", payload.ErrorMsg, "By:", thePatient.Id)
					tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
					return
				}

				payload.Appt.editAppointment(t, payload.Appt.Patient, payload.Appt.Doctor)
				session.SetNotification(req, "Appointment updated!", "Success")
				Info.Println(req.RemoteAddr, " Appointment updated successfully:", payload.Appt.Id, "By:", thePatient.Id)
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
				Info.Println(req.RemoteAddr, " [Admin] User deleted successfully. By:", thePatient.Id)
			} else {
				payload.ErrorMsg = ErrPatientIDNotFound.Error()
				Error.Println(req.RemoteAddr, "[Admin] User deletion failure:", payload.ErrorMsg, "By:", thePatient.Id)
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
					session.SetNotification(req, "Error adding to payment queue! Appointment ID not found.", "Error")
					Error.Println(req.RemoteAddr, "[Admin] Payment enqueue failure: Appointment ID not found. By:", thePatient.Id)
					http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
					return
				}

				appt := appointments[apptIdIndex]
				createPayment(appt, 19.99)
				session.SetNotification(req, "Appointment added to payment queue!", "Success")
				Info.Println(req.RemoteAddr, " [Admin] Payment enqueued successfully. By:", thePatient.Id)
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
		if paymentQ.Front != nil {
			Info.Println(req.RemoteAddr, " [Admin] Payment dequeued successfully. Appt:", paymentQ.Front.Payment.Appointment.Id, "By:", thePatient.Id)
			paymentQ.dequeue()
		}
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
		if paymentQ.Front != nil {
			Info.Println(req.RemoteAddr, " [Admin] Payment dequeued to missed queue successfully. Appt:", paymentQ.Front.Payment.Appointment.Id, "By:", thePatient.Id)
			paymentQ.dequeueToMissedPaymentQueue()
		}
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
		if missedPaymentQ.Front != nil {
			Info.Println(req.RemoteAddr, " [Admin] Payment dequeued to main queue successfully. Appt:", missedPaymentQ.Front.Payment.Appointment.Id, "By:", thePatient.Id)
			missedPaymentQ.dequeueToPaymentQueue()
		}
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
	fmt.Println("Sessions:", len(session.MapSessions), session.MapSessions)
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
