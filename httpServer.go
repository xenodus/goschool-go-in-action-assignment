package main

import (
	"fmt"
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

	// Doctor - Handlers in doctor.go
	http.HandleFunc(pageDoctors, viewDoctorsPage)

	// Admins - Handlers in patient.go
	http.HandleFunc(pageAdminAllAppointments, adminAppointmentPage)
	http.HandleFunc(pageAdminEditAppointment, adminEditAppointmentPage)
	http.HandleFunc(pageAdminSessions, adminSessionsPage)
	http.HandleFunc(pageAdminUsers, adminUsersPage)

	http.HandleFunc(pageAdminDebug, adminDebugPage)

	http.HandleFunc(pageAdminPaymentEnqueue, adminPaymentEnqueuePage)
	http.HandleFunc(pageAdminPaymentDequeue, adminPaymentDequeuePage)
	http.HandleFunc(pageAdminPaymentDequeueToMissedQueue, adminPaymentDequeueToMissedQueuePage)
	http.HandleFunc(pageAdminPaymentDequeueToPaymentQueue, adminPaymentDequeueToPaymentQueuePage)

	// Payment Queue - Handlers in payment.go
	http.HandleFunc(pagePaymentQueue, paymentQueuePage)

	fmt.Println("Server starting on: http://" + serverHost + ":" + serverPort)

	err := http.ListenAndServe(serverHost+":"+serverPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
