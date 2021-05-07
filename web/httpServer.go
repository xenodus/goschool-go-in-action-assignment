package web

import (
	"assignment4/clinic"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

// Parse directory for templates and add helper functions to be used inside.
func init() {

	funcMap := template.FuncMap{
		"time2HumanReadable":     time2HumanReadable,
		"time2HumanReadableFull": time2HumanReadableFull,
		"getUserByID":            getUserByID,
		"ucFirst":                ucFirst,
		"stripSpace":             stripSpace,
		"maxAdvanceApptDays":     maxAdvanceApptDays,
	}

	tpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*"))
}

// StartHttpServer setup routes & handlers and start web server over https.
// It also calls SeedData from clinic package to perform database setup and/or seeding of test data depending on clinic package's config settings.
func StartHttpServer(myDb *sql.DB) {

	clinic.SetDb(myDb)
	clinic.SeedData()

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Used in IndexPage
	http.HandleFunc(pageError, genericErrorHandler)

	// Index
	http.HandleFunc(pageIndex, indexPage)

	// Appointments - Handlers in appointment.go
	http.HandleFunc(pageMyAppointments, appointmentPage)
	http.HandleFunc(pageNewAppointment, newAppointmentPage)
	http.HandleFunc(pageEditAppointment, editAppointmentPage)

	// User - Handlers in patient.go
	http.HandleFunc(pageLogin, loginPage)
	http.HandleFunc(pageLogout, logoutPage)
	http.HandleFunc(pageRegister, registerPage)
	http.HandleFunc(pageProfile, profilePage)

	// Doctor - Handlers in doctor.go
	http.HandleFunc(pageDoctors, viewDoctorsPage)

	// Admins - Handlers in patient.go
	http.HandleFunc(pageAdminAllAppointments, adminAppointmentPage)
	http.HandleFunc(pageAdminEditAppointment, adminEditAppointmentPage)
	http.HandleFunc(pageAdminSessions, adminSessionsPage)
	http.HandleFunc(pageAdminUsers, adminUsersPage)

	http.HandleFunc(pageAdminPaymentEnqueue, adminPaymentEnqueuePage)
	http.HandleFunc(pageAdminPaymentDequeue, adminPaymentDequeuePage)
	http.HandleFunc(pageAdminPaymentDequeueToMissedQueue, adminPaymentDequeueToMissedQueuePage)
	http.HandleFunc(pageAdminPaymentDequeueToPaymentQueue, adminPaymentDequeueToPaymentQueuePage)

	// Payment Queue - Handlers in payment.go
	http.HandleFunc(pagePaymentQueue, paymentQueuePage)

	// PSI
	http.HandleFunc(pagePSI, psiPage)

	// Debug Page
	http.HandleFunc(pageAdminDebug, adminDebugPage)

	fmt.Println("Server starting on: https://" + serverHost + ":" + serverPort)

	err := http.ListenAndServeTLS(serverHost+":"+serverPort, "./.cert/.cert.pem", "./.cert/.key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
