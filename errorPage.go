package main

import "net/http"

func errorPage(res http.ResponseWriter, req *http.Request) {

	// Anonymous payload
	payload := struct {
		PageTitle  string
		ErrorMsg   string
		SuccessMsg string
		User       *patient
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
		payload.ErrorMsg = ErrInternalServerError.Error()
	}

	res.WriteHeader(errorCode)
	tpl.ExecuteTemplate(res, "error.gohtml", payload)
}
