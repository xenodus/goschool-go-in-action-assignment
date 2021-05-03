package main

import (
	"net/http"

	"assignment4/psi"
)

func psiPage(res http.ResponseWriter, req *http.Request) {

	thePatient, _ := isLoggedIn(req)

	// Anonymous payload
	payload := struct {
		User           *patient
		PageTitle      string
		ErrorMsg       string
		Psi            string
		PsiDescription string
	}{
		thePatient, "PSI", "", "", "",
	}

	psi, err := psi.GetPSI()

	if err != nil {
		payload.ErrorMsg = "Unable to retrieve PSI"
		Error.Println(req.RemoteAddr, "[Admin]", err.Error()) // only show detailed error inside logs
	} else {
		payload.Psi = psi.Value
		payload.PsiDescription = psi.Description
	}

	tpl.ExecuteTemplate(res, "psi.gohtml", payload)
}
