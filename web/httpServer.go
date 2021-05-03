package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func init() {

	// Adding helper functions to templates
	funcMap := template.FuncMap{
		"time2HumanReadable": time2HumanReadable,
		"getUserByID":        getUserByID,
		"ucFirst":            ucFirst,
		"stripSpace":         stripSpace,
	}

	tpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*"))
}

func StartHttpServer() {
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
