package web

import (
	"assignment4/clinic"
	"net/http"
)

func indexPage(res http.ResponseWriter, req *http.Request) {

	if req.URL.Path != pageIndex {
		notFoundErrorHandler(res, req)
		return
	}

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *clinic.Patient
	}{
		"",
		thePatient,
	}

	tpl.ExecuteTemplate(res, "index.gohtml", payload)
}
