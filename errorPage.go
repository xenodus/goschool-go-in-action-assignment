package main

import "net/http"

func errorPage(res http.ResponseWriter, req *http.Request) {

	err := req.FormValue("err")

	var errorMsg = ErrInternalServerError.Error()
	var errorCode = http.StatusInternalServerError

	switch err {
	default:
		errorCode = http.StatusInternalServerError
		errorMsg = ErrInternalServerError.Error()
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		ErrorMsg  string
		User      *patient
	}{
		"Error",
		errorMsg,
		nil,
	}

	res.WriteHeader(errorCode)
	tpl.ExecuteTemplate(res, "error.gohtml", payload)
}
