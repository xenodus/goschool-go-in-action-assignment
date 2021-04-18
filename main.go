// Author: Alvin Yeoh

package main

import (
	"html/template"
	"log"
	"net/http"
	"runtime"
	"sync"
)

// Globals
var doctors = []*doctor{}
var patients = []*patient{}
var appointments = []*appointment{}

var paymentQ = paymentQueue{}
var missedPaymentQ = paymentQueue{}
var appointmentsSortedByTimeslot = []*appointment{}

var admins = []string{}

var doctorsBST *BST
var wg sync.WaitGroup

var tpl *template.Template
var mapSessions = map[string]string{}

var cookieID string

func init() {
	// Essentials
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

func startHttpServer() {
	//mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Index
	http.HandleFunc(pageIndex, index)

	// Appointments - Handlers in appointment.go
	http.HandleFunc(pageMyAppointments, appointmentPage)
	http.HandleFunc(pageNewAppointment, newAppointmentPage)
	http.HandleFunc(pageEditAppointment, editAppointmentPage)

	// User - Handlers in patient.go
	http.HandleFunc(pageLogin, login)
	http.HandleFunc(pageLogout, logout)
	http.HandleFunc(pageRegister, register)
	http.HandleFunc(pageProfile, profile)

	// Doctor
	http.HandleFunc(pageDoctors, viewDoctorsPage)

	// Admins
	http.HandleFunc(pageAdminAllAppointments, adminAppointmentPage)
	http.HandleFunc(pageAdminSessions, adminSessionsPage)
	http.HandleFunc(pageAdminUsers, adminUsersPage)

	http.HandleFunc(pageAdminPaymentEnqueue, adminPaymentEnqueuePage)
	http.HandleFunc(pageAdminPaymentDequeue, paymentDequeuePage)
	http.HandleFunc(pageAdminPaymentDequeueToMissedQueue, paymentDequeueToMissedQueuePage)

	// Payment Queue
	http.HandleFunc(pagePaymentQueue, paymentQueuePage)

	http.Handle("/favicon.ico", http.NotFoundHandler())

	err := http.ListenAndServe(":5221", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Public Web Pages
func index(res http.ResponseWriter, req *http.Request) {

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
