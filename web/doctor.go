package web

import (
	"assignment4/clinic"
	"net/http"
	"strconv"
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
	}{
		"View Doctors",
		"",
		thePatient,
		nil,
		nil,
		clinic.Doctors,
	}

	// Get querystring values
	doctorID := req.FormValue("doctorID")

	if doctorID != "" {
		doctorID, _ := strconv.ParseInt(doctorID, 10, 64)
		doc, err := clinic.DoctorsBST.GetDoctorByIDBST(doctorID)

		if err == nil {
			payload.ChosenDoctor = doc
			payload.TimeslotsAvailable = clinic.GetAvailableTimeslot(payload.ChosenDoctor.Appointments)
		} else {
			payload.ErrorMsg = err.Error()
			doLog(req, "ERROR", " Doctor lookup failure: "+payload.ErrorMsg+" ID: "+strconv.FormatInt(doctorID, 10))
		}
	}

	tpl.ExecuteTemplate(res, "doctors.gohtml", payload)
}
