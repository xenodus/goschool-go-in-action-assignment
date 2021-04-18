package main

import (
	"log"
	"net/http"
)

func startHttpServer() {
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc(pageError, errorPage)

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

	// Doctor
	http.HandleFunc(pageDoctors, viewDoctorsPage)

	// Admins
	http.HandleFunc(pageAdminAllAppointments, adminAppointmentPage)
	http.HandleFunc(pageAdminSessions, adminSessionsPage)
	http.HandleFunc(pageAdminUsers, adminUsersPage)

	http.HandleFunc(pageAdminPaymentEnqueue, adminPaymentEnqueuePage)
	http.HandleFunc(pageAdminPaymentDequeue, adminPaymentDequeuePage)
	http.HandleFunc(pageAdminPaymentDequeueToMissedQueue, adminPaymentDequeueToMissedQueuePage)
	http.HandleFunc(pageAdminPaymentDequeueToPaymentQueue, adminPaymentDequeueToPaymentQueuePage)

	// Payment Queue
	http.HandleFunc(pagePaymentQueue, paymentQueuePage)

	http.Handle("/favicon.ico", http.NotFoundHandler())

	err := http.ListenAndServe(":5221", nil)
	if err != nil {
		log.Fatal(err)
	}
}
