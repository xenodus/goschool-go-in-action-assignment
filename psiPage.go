package main

import (
	"net/http"

	"./internal/psi"
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
		payload.ErrorMsg = err.Error()
		Error.Println(req.RemoteAddr, "[Admin]", payload.ErrorMsg)
		tpl.ExecuteTemplate(res, "psi.gohtml", payload)
		return
	}

	payload.Psi = psi.Value
	payload.PsiDescription = psi.Description

	tpl.ExecuteTemplate(res, "psi.gohtml", payload)
}
