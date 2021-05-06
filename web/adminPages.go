package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"assignment4/clinic"
	"assignment4/session"
)

var tpl *template.Template

func adminSessionsPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

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
		User       *clinic.Patient
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
		action := strings.ToLower(req.FormValue("action"))
		sessionId := req.FormValue("sessionId")

		// delete single session
		if action == "delete" && sessionId != "" {
			if _, ok := session.MapSessions[sessionId]; ok {
				delete(session.MapSessions, sessionId)
				payload.SuccessMsg = "Session deleted!"
				go doLog(req, "INFO", " [Admin] Session deleted successfully. By: "+thePatient.Id)
			} else {
				payload.ErrorMsg = errSessionNotFound.Error()
				go doLog(req, "ERROR", "[Admin] Session delete failure: "+payload.ErrorMsg+" By: "+thePatient.Id)
			}
		}

		// delete all sessions
		if action == "purge" {
			session.MapSessions = make(map[string]session.Session)
		}

		if action != "" {
			// if user's session is gone, redirect away
			if _, isLoggedInCheck := clinic.IsLoggedIn(req); !isLoggedInCheck {
				http.Redirect(res, req, pageLogin, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "adminSessions.gohtml", payload)
}

func adminAppointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

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
		User         *clinic.Patient
		Appointments []*clinic.Appointment
		ErrorMsg     string
		SuccessMsg   string
	}{
		"Manage Appointments",
		thePatient,
		clinic.AppointmentsSortedByTimeslot,
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

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

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
		User               *clinic.Patient
		Appt               *clinic.Appointment
		TimeslotsAvailable []int64
		ErrorMsg           string
		ChosenDate         string
	}{
		"Edit Appointment", thePatient, nil, nil, "", "",
	}

	// Get querystring values
	inputApptId := req.FormValue("apptId")
	action := strings.ToLower(req.FormValue("action"))

	if action != "edit" && action != "cancel" {
		go doLog(req, "ERROR", "[Admin] Appointment update failure: invalid action type. By: "+thePatient.Id)
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	apptId, err := strconv.ParseInt(inputApptId, 10, 64)

	if err != nil {
		session.SetNotification(req, clinic.ErrAppointmentIDNotFound.Error(), "Error")
		go doLog(req, "ERROR", "[Admin] Appointment update failure: invalid appt id. Unable to parse. By: "+thePatient.Id)
		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	// Check if appt id is valid
	theApptIndex := clinic.BinarySearchApptID(apptId)

	if theApptIndex < 0 {
		session.SetNotification(req, clinic.ErrAppointmentIDNotFound.Error(), "Error")
		go doLog(req, "ERROR", "[Admin] Appointment update failure: "+clinic.ErrAppointmentIDNotFound.Error())

		fmt.Println("all clinic appts...")
		for _, v := range clinic.Appointments {
			fmt.Println(v.Id)
		}

		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	payload.Appt = clinic.Appointments[theApptIndex]

	// Cancel Appt
	if action == "cancel" {
		if req.Method == http.MethodPost {
			session.SetNotification(req, "Appointment cancelled!", "Success")
			go doLog(req, "INFO", "[Admin] Appointment cancelled successfully: "+strconv.FormatInt(payload.Appt.Id, 10)+" By: "+thePatient.Id)
			payload.Appt.CancelAppointment()
		} else {
			go doLog(req, "ERROR", "[Admin] Appointment cancellation failure: GET REQUEST. By:"+thePatient.Id)
		}

		http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
		return
	}

	// Edit Appt
	if action == "edit" {

		if req.Method == http.MethodPost {

			date := req.FormValue("date")
			timeslot := req.FormValue("timeslot")

			// Date
			if date != "" {
				// Parse date
				dt, dtErr := time.Parse("02 January 2006", date)

				if dtErr != nil {
					payload.ErrorMsg = "Invalid date"
					go doLog(req, "ERROR", "[Admin] Appointment update failure: "+payload.ErrorMsg+dtErr.Error())
					tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
					return
				}

				payload.ChosenDate = date
				payload.TimeslotsAvailable = clinic.GetAvailableTimeslot(dt.Unix(), append(payload.Appt.Doctor.GetAppointmentsByDate(dt.Unix()), payload.Appt.Patient.GetAppointmentsByDate(dt.Unix())...))
				_, timeSlotErr := clinic.IsThereTimeslot(dt.Unix(), payload.Appt.Patient, payload.Appt.Doctor)

				if timeSlotErr != nil {
					payload.ErrorMsg = timeSlotErr.Error()
					go doLog(req, "ERROR", "[Admin] Appointment update failure: "+payload.ErrorMsg+" By: "+thePatient.Id)
					tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
					return
				}

				if timeslot != "" {

					t, _ := strconv.ParseInt(timeslot, 10, 64)

					_, isApptTimeValidErr := clinic.IsApptTimeValid(t) // Is time in the past check

					// Past time
					if isApptTimeValidErr != nil {
						payload.ErrorMsg = isApptTimeValidErr.Error()
						go doLog(req, "ERROR", "[Admin] Appointment update failure: "+payload.ErrorMsg+" By: "+thePatient.Id)
						tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
						return
					}

					// Patient / Doctor time check
					if !payload.Appt.Patient.IsFreeAt(t) || !payload.Appt.Doctor.IsFreeAt(t) {
						payload.ErrorMsg = clinic.ErrDuplicateTimeslot.Error()
						go doLog(req, "ERROR", "[Admin] Appointment update failure: "+payload.ErrorMsg+" By: "+thePatient.Id)
						tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
						return
					}

					payload.Appt.EditAppointment(t, payload.Appt.Patient, payload.Appt.Doctor)
					session.SetNotification(req, "Appointment updated!", "Success")
					go doLog(req, "INFO", "[Admin] Appointment updated successfully: "+strconv.FormatInt(payload.Appt.Id, 10)+" By: "+thePatient.Id)
					http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
					return
				}
			}
		}
	}

	tpl.ExecuteTemplate(res, "adminEditAppointment.gohtml", payload)
}

func adminUsersPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

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
		User       *clinic.Patient
		Patients   []*clinic.Patient
		ErrorMsg   string
		SuccessMsg string
	}{
		"Manage Users",
		thePatient,
		clinic.Patients,
		"",
		"",
	}

	if req.Method == http.MethodPost {
		// Get querystring values
		action := strings.ToLower(req.FormValue("action"))
		userId := req.FormValue("userId")

		// delete single users
		if action == "delete" && userId != "" {
			theUser, err := clinic.GetPatientByID(userId)

			if err == nil {
				theUser.DeletePatient()
				payload.Patients = clinic.Patients
				payload.SuccessMsg = "User deleted!"
				go doLog(req, "INFO", "[Admin] User deleted successfully. By: "+thePatient.Id)
			} else {
				payload.ErrorMsg = clinic.ErrPatientIDNotFound.Error()
				go doLog(req, "ERROR", "[Admin] User deletion failure: "+payload.ErrorMsg+" By: "+thePatient.Id)
			}
		}

		if action != "" {
			// if self is gone, bye bye
			if _, isLoggedInCheck := clinic.IsLoggedIn(req); !isLoggedInCheck {
				http.Redirect(res, req, pageLogin, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "adminUsers.gohtml", payload)
}

func adminPaymentEnqueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

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
				apptIdIndex := clinic.BinarySearchApptID(apptId)

				if apptIdIndex < 0 {
					session.SetNotification(req, "Error adding to payment queue! Appointment ID not found.", "Error")
					go doLog(req, "ERROR", "[Admin] Payment enqueue failure: Appointment ID not found. By: "+thePatient.Id)
					http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
					return
				}

				appt := clinic.Appointments[apptIdIndex]
				clinic.CreatePayment(appt, 19.99, nil)
				appt.CancelAppointment()
				session.SetNotification(req, "Appointment added to payment queue!", "Success")
				go doLog(req, "INFO", "[Admin] Payment enqueued successfully. By: "+thePatient.Id)
				http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
				return
			}
		}
	}

	http.Redirect(res, req, pageAdminAllAppointments, http.StatusSeeOther)
}

func adminPaymentDequeuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		if clinic.PaymentQ.Front != nil {
			go doLog(req, "INFO", "[Admin] Payment dequeued successfully. Appt: "+strconv.FormatInt(clinic.PaymentQ.Front.Payment.Appointment.Id, 10)+" By: "+thePatient.Id)
			clinic.PaymentQ.Front.Payment.ClearPayment()
			clinic.PaymentQ.Dequeue()
		}
	}

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminPaymentDequeueToMissedQueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		if clinic.PaymentQ.Front != nil {
			go doLog(req, "INFO", "[Admin] Payment dequeued to missed queue successfully. Appt: "+strconv.FormatInt(clinic.PaymentQ.Front.Payment.Appointment.Id, 10)+" By: "+thePatient.Id)
			clinic.PaymentQ.DequeueToMissedPaymentQueue()
		}
	}

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminPaymentDequeueToPaymentQueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	if !thePatient.IsAdmin() {
		http.Error(res, "Restricted Zone", http.StatusUnauthorized)
		return
	}

	if req.Method == http.MethodPost {
		if clinic.MissedPaymentQ.Front != nil {
			go doLog(req, "INFO", "[Admin] Payment dequeued to main queue successfully. Appt: "+strconv.FormatInt(clinic.MissedPaymentQ.Front.Payment.Appointment.Id, 10)+" By: "+thePatient.Id)
			clinic.MissedPaymentQ.DequeueToPaymentQueue()
		}
	}

	http.Redirect(res, req, pagePaymentQueue, http.StatusSeeOther)
}

func adminDebugPage(res http.ResponseWriter, req *http.Request) {

	fmt.Println(":::::::::::::: Debug Dump ::::::::::::::")
	fmt.Println("======================================")
	fmt.Println("Appointments:", len(clinic.Appointments), clinic.Appointments)
	fmt.Println("Appoinments sorted by time:", len(clinic.AppointmentsSortedByTimeslot), clinic.AppointmentsSortedByTimeslot)
	fmt.Println("Doctors:", len(clinic.Doctors), clinic.Doctors)
	fmt.Println("Doctors BST:", clinic.DoctorsBST)
	fmt.Println("Patients:", len(clinic.Patients), clinic.Patients)
	fmt.Println("PaymentQueue:", clinic.PaymentQ.Size, clinic.PaymentQ)
	fmt.Println("MissedPaymentQueue:", clinic.MissedPaymentQ.Size, clinic.MissedPaymentQ)
	fmt.Println("Sessions:", len(session.MapSessions), session.MapSessions)
	fmt.Println("Admins:", len(clinic.Admins), clinic.Admins)

	fmt.Println(":::::::::::::: Appointments ::::::::::::::")
	fmt.Println("--- Id | Appt Time :::", len(clinic.Appointments))

	for _, v := range clinic.Appointments {
		fmt.Println(v.Id, time2HumanReadable(v.Time))
	}

	fmt.Println(":::::::::::::: Appointments By Time ::::::::::::::")
	fmt.Println("--- Id | Appt Time :::", len(clinic.AppointmentsSortedByTimeslot))

	for _, v := range clinic.AppointmentsSortedByTimeslot {
		fmt.Println(v.Id, time2HumanReadable(v.Time))
	}

	fmt.Println(":::::::::::::: Doctors ::::::::::::::")
	fmt.Println("--- Id | # of Appts :::", len(clinic.Doctors))

	for _, v := range clinic.Doctors {
		fmt.Println(v.Id, len(v.Appointments))
	}

	fmt.Println(":::::::::::::: Patients ::::::::::::::")
	fmt.Println("--- Id | # of Appts :::", len(clinic.Patients))

	for _, v := range clinic.Patients {
		fmt.Println(v.Id, len(v.Appointments))
	}

	http.Redirect(res, req, pageIndex, http.StatusSeeOther)
}
