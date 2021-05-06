package web

import (
	"assignment4/clinic"
	"net/http"
	"strconv"
	"time"
)

func viewDoctorsPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle          string
		ErrorMsg           string
		User               *clinic.Patient
		ChosenDoctor       *clinic.Doctor
		TimeslotsAvailable []int64
		Doctors            []*clinic.Doctor
		ChosenDate         string
	}{
		"View Doctors",
		"",
		thePatient,
		nil,
		nil,
		clinic.Doctors,
		"",
	}

	if req.Method == http.MethodPost {

		doctorID := req.FormValue("doctorID")
		date := req.FormValue("date")

		if doctorID != "" {
			doctorID, _ := strconv.ParseInt(doctorID, 10, 64)
			doc, docErr := clinic.DoctorsBST.GetDoctorByIDBST(doctorID)

			if docErr != nil {
				payload.ErrorMsg = docErr.Error()
				go doLog(req, "ERROR", " Doctor lookup failure: "+payload.ErrorMsg+" ID: "+strconv.FormatInt(doctorID, 10))
				tpl.ExecuteTemplate(res, "doctors.gohtml", payload)
				return
			}

			payload.ChosenDoctor = doc
		}

		if date != "" {
			dt, dtErr := time.Parse("02 January 2006", date)

			if dtErr != nil {
				payload.ErrorMsg = "Invalid date"
				go doLog(req, "ERROR", " Doctor lookup failure: "+payload.ErrorMsg)
				tpl.ExecuteTemplate(res, "doctors.gohtml", payload)
				return
			}

			payload.ChosenDate = date
			payload.TimeslotsAvailable = clinic.GetAvailableTimeslot(dt.Unix(), payload.ChosenDoctor.GetAppointmentsByDate(dt.Unix()))
		}
	}

	tpl.ExecuteTemplate(res, "doctors.gohtml", payload)
}
