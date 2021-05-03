package web

import (
	"assignment4/clinic"
	"net/http"
)

func genericErrorHandler(res http.ResponseWriter, req *http.Request) {

	// Anonymous payload
	payload := struct {
		PageTitle  string
		ErrorMsg   string
		SuccessMsg string
		User       *clinic.Patient
	}{
		"Error",
		"",
		"",
		nil,
	}

	err := req.FormValue("err")
	var errorCode = http.StatusInternalServerError

	switch err {
	// other cases
	// default
	default:
		errorCode = http.StatusInternalServerError
		payload.ErrorMsg = errInternalServerError.Error()
	}

	res.WriteHeader(errorCode)
	tpl.ExecuteTemplate(res, "error.gohtml", payload)
}

func notFoundErrorHandler(res http.ResponseWriter, req *http.Request) {
	// Anonymous payload
	payload := struct {
		PageTitle string
		ErrorMsg  string
		User      *clinic.Patient
	}{
		"Page not found",
		errStatusNotFound.Error(),
		nil,
	}

	res.WriteHeader(http.StatusNotFound)
	tpl.ExecuteTemplate(res, "404.gohtml", payload)
}
