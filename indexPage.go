package main

import "net/http"

func indexPage(res http.ResponseWriter, req *http.Request) {

	if req.URL.Path != pageIndex {
		notFoundErrorHandler(res, req)
		return
	}

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
	}{
		"",
		thePatient,
	}

	tpl.ExecuteTemplate(res, "index.gohtml", payload)
}
