package main

import "net/http"

func indexPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	// Anonymous payload
	payload := struct {
		User *patient
	}{
		thePatient,
	}

	tpl.ExecuteTemplate(res, "index.gohtml", payload)
}
