package web

import (
	"net/http"

	"assignment4/clinic"
	"assignment4/psi"
)

func psiPage(res http.ResponseWriter, req *http.Request) {

	thePatient, _ := clinic.IsLoggedIn(req)

	// Anonymous payload
	payload := struct {
		User           *clinic.Patient
		PageTitle      string
		ErrorMsg       string
		Psi            string
		PsiDescription string
	}{
		thePatient, "PSI", "", "", "",
	}

	thePsi, err := psi.GetPSI()

	if err != nil {
		payload.ErrorMsg = "Unable to retrieve PSI"
		go doLog(req, "ERROR", "[Admin] "+err.Error()) // only show detailed error inside logs
	} else {
		payload.Psi = thePsi.Value
		payload.PsiDescription = thePsi.Description
	}

	tpl.ExecuteTemplate(res, "psi.gohtml", payload)
}
