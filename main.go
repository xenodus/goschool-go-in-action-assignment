// Author: Alvin Yeoh

package main

import (
	"html/template"
	"net/http"
	"runtime"
	"sync"
)

// Globals
var doctors = []*doctor{}
var patients = []*patient{}
var appointments = []*appointment{}
var appointmentsSortedByTimeslot = []*appointment{}

var paymentQ = paymentQueue{}
var missedPaymentQ = paymentQueue{}

var admins = []string{}

var doctorsBST *BST
var wg sync.WaitGroup
var mutex sync.Mutex

var tpl *template.Template
var mapSessions = map[string]string{}

var cookieID string

func init() {
	// Essentials Test Data
	seedDoctors()
	seedAdmins()
	seedPatients()

	// Just randomizing the cookie name on each init
	cookieID = getRandomCookiePrefix()

	// Adding helper functions to templates
	funcMap := template.FuncMap{
		"time2HumanReadable": time2HumanReadable,
		"isUserAdminByID":    isUserAdminByID,
	}

	tpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*"))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	startHttpServer()
}

func errorPage(res http.ResponseWriter, req *http.Request) {

	err := req.FormValue("err")

	var errorMsg = ErrInternalServerError.Error()
	var errorCode = http.StatusInternalServerError

	switch err {
	case "ErrInternalServerError":
		errorCode = http.StatusInternalServerError
		errorMsg = ErrInternalServerError.Error()
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
