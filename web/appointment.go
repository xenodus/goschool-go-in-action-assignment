package web

import (
	"assignment4/clinic"
	"assignment4/session"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func editAppointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Get querystring values
	inputApptId := req.FormValue("apptId")
	action := strings.ToLower(req.FormValue("action"))

	if action != "edit" && action != "cancel" {
		go doLog(req, "ERROR", " Appointment update failure: invalid action type")
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle          string
		User               *clinic.Patient
		Appt               *clinic.Appointment
		TimeslotsAvailable []int64
		ErrorMsg           string
		SuccessMsg         string
		ChosenDate         string
	}{
		"Edit Appointment", thePatient, nil, nil, "", "", "",
	}

	apptId, err := strconv.ParseInt(inputApptId, 10, 64)

	if err != nil {
		session.SetNotification(req, clinic.ErrAppointmentIDNotFound.Error(), "Error")
		go doLog(req, "ERROR", " Appointment update failure: invalid appt id. Unable to parse.")
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Check if appt id is valid
	patientApptIDIndex := clinic.BinarySearchApptID(apptId)

	if patientApptIDIndex < 0 {
		session.SetNotification(req, clinic.ErrAppointmentIDNotFound.Error(), "Error")
		go doLog(req, "ERROR", " Appointment update failure:"+clinic.ErrAppointmentIDNotFound.Error())
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	payload.Appt = clinic.Appointments[patientApptIDIndex]

	// Does not belong to logged in user
	if payload.Appt.Patient != thePatient {
		session.SetNotification(req, clinic.ErrAppointmentIDNotFound.Error(), "Error")
		go doLog(req, "ERROR", " Appointment update failure: appt "+strconv.FormatInt(payload.Appt.Id, 10)+" does not belong to user "+thePatient.Id)
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Cancel Appt
	if action == "cancel" {
		if req.Method == http.MethodPost {
			session.SetNotification(req, "Appointment cancelled!", "Success")
			go doLog(req, "INFO", " Appointment cancelled successfully: "+strconv.FormatInt(payload.Appt.Id, 10))
			payload.Appt.CancelAppointment()
		} else {
			go doLog(req, "ERROR", " Appointment cancellation failure: GET REQUEST")
		}

		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Edit Appt
	if action == "edit" {

		if req.Method == http.MethodPost {

			date := req.FormValue("date")
			timeslot := req.FormValue("timeslot")

			if date != "" {
				// Parse date
				dt, dtErr := time.Parse("02 January 2006", date)

				if dtErr != nil {
					payload.ErrorMsg = "Invalid date"
					go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg+dtErr.Error())
					tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
					return
				}

				payload.ChosenDate = date
				payload.TimeslotsAvailable = clinic.GetAvailableTimeslot(dt.Unix(), append(payload.Appt.Doctor.GetAppointmentsByDate(dt.Unix()), payload.User.GetAppointmentsByDate(dt.Unix())...))
				_, timeSlotErr := clinic.IsThereTimeslot(dt.Unix(), payload.User, payload.Appt.Doctor)

				if timeSlotErr != nil {
					payload.ErrorMsg = clinic.ErrNoMoreTimeslot.Error()
					go doLog(req, "ERROR", " Appointment update failure: "+payload.ErrorMsg)
					tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
					return
				}

				// Form submit values
				if timeslot != "" {

					t, timeErr := strconv.ParseInt(timeslot, 10, 64)

					if timeErr != nil {
						payload.ErrorMsg = "Invalid timeslot"
						go doLog(req, "ERROR", "Appointment update failure: "+payload.ErrorMsg+timeErr.Error())
						tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
						return
					}

					// Patient / Doctor time check
					if !payload.Appt.Patient.IsFreeAt(t) || !payload.Appt.Doctor.IsFreeAt(t) {
						payload.ErrorMsg = clinic.ErrDuplicateTimeslot.Error()
						go doLog(req, "ERROR", " Appointment update failure: "+payload.ErrorMsg)
						tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
						return
					}

					_, isApptTimeValidErr := clinic.IsApptTimeValid(t)

					// Past time
					if isApptTimeValidErr != nil {
						payload.ErrorMsg = isApptTimeValidErr.Error()
						go doLog(req, "ERROR", " Appointment update failure: "+payload.ErrorMsg)
						tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
						return
					}

					payload.Appt.EditAppointment(t, payload.Appt.Patient, payload.Appt.Doctor)
					session.SetNotification(req, "Appointment updated!", "Success")
					go doLog(req, "INFO", " Appointment updated successfully:"+strconv.FormatInt(payload.Appt.Id, 10))
					http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
					return
				}
			}
		}
	}

	tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
}

func appointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle  string
		User       *clinic.Patient
		SuccessMsg string
		ErrorMsg   string
	}{
		"My Appointments", thePatient, "", "",
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

	tpl.ExecuteTemplate(res, "appointments.gohtml", payload)
}

func newAppointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Form submit values
	doctorID := req.FormValue("doctor")
	date := req.FormValue("date")
	timeslot := req.FormValue("timeslot")

	payload := struct {
		PageTitle          string
		User               *clinic.Patient
		Doctors            []*clinic.Doctor
		ChosenDoctor       *clinic.Doctor
		TimeslotsAvailable []int64
		ErrorMsg           string
		ChosenDate         string
	}{
		"New Appointment", thePatient, clinic.Doctors, nil, nil, "", "",
	}

	if req.Method == http.MethodPost {

		if doctorID != "" {
			doctorID, _ := strconv.ParseInt(doctorID, 10, 64)
			doc, err := clinic.DoctorsBST.GetDoctorByIDBST(doctorID)

			if err != nil {
				payload.ErrorMsg = err.Error()
				go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg)
				tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
				return
			}

			payload.ChosenDoctor = doc

			if date != "" {
				// Parse date
				dt, dtErr := time.Parse("02 January 2006", date)

				if dtErr != nil {
					payload.ErrorMsg = "Invalid date"
					go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg+dtErr.Error())
					tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
					return
				}

				payload.ChosenDate = date
				payload.TimeslotsAvailable = clinic.GetAvailableTimeslot(dt.Unix(), append(payload.ChosenDoctor.GetAppointmentsByDate(dt.Unix()), thePatient.GetAppointmentsByDate(dt.Unix())...))
				_, timeSlotErr := clinic.IsThereTimeslot(dt.Unix(), thePatient, payload.ChosenDoctor)

				if timeSlotErr != nil {
					if timeSlotErr == clinic.ErrDoctorNoMoreTimeslot {
						payload.ErrorMsg = "Dr. " + payload.ChosenDoctor.First_name + " " + payload.ChosenDoctor.Last_name + " has no more available timeslots for " + date
					} else if timeSlotErr == clinic.ErrPatientNoMoreTimeslot {
						payload.ErrorMsg = "You have no more available timeslots for " + date
					} else {
						payload.ErrorMsg = timeSlotErr.Error()
					}

					payload.ChosenDoctor = nil

					go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg)
					tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
					return
				}

				if timeslot != "" {
					t, timeErr := strconv.ParseInt(timeslot, 10, 64)

					if timeErr != nil {
						payload.ErrorMsg = "Invalid timeslot"
						go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg+timeErr.Error())
						tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
						return
					}

					// Check if slot truely exists
					if !payload.ChosenDoctor.IsFreeAt(t) || !thePatient.IsFreeAt(t) {
						payload.ErrorMsg = clinic.ErrDuplicateTimeslot.Error()
						go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg)
						tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
						return
					}

					_, isApptTimeValidErr := clinic.IsApptTimeValid(t)

					// Past time
					if isApptTimeValidErr != nil {
						payload.ErrorMsg = isApptTimeValidErr.Error()
						go doLog(req, "ERROR", "Appointment creation failure: "+payload.ErrorMsg)
						tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
						return
					}

					newAppt, newApptErr := clinic.MakeAppointment(t, thePatient, payload.ChosenDoctor, nil)

					if newApptErr == nil {
						session.SetNotification(req, "Appointment scheduled!", "Success")
						go doLog(req, "INFO", "Appointment created successfully:"+strconv.FormatInt(newAppt.Id, 10))
						http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
						return
					}
				}
			}
		}
	}

	tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
}
