package main

import "net/http"

func indexPage(res http.ResponseWriter, req *http.Request) {

	if req.URL.Path != pageIndex {
		notFoundErrorHandler(res, req)
		return
	}

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

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

func notFoundErrorHandler(res http.ResponseWriter, req *http.Request) {
	// Anonymous payload
	payload := struct {
		PageTitle string
		ErrorMsg  string
		User      *patient
	}{
		"Page not found",
		ErrStatusNotFound.Error(),
		nil,
	}

	res.WriteHeader(http.StatusNotFound)
	tpl.ExecuteTemplate(res, "404.gohtml", payload)
}
