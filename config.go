package main

// For doctors' timeslots - 1st consultation @ 8 am, last @ 10 pm
const startOperationHour = 8
const endOperationHour = 22
const appointmentIntervals = 30 // 30 mins between each consultations

// Disabled for ease of testing of assignment
const strictNRIC = false

// Set current hour minute for testing
const testFakeTime = false
const testHour = 9
const testMinute = 15

// URLS
const pageIndex = "/"
const pageLogin = "/login"
const pageLogout = "/logout"
const pageRegister = "/register"
const pageProfile = "/profile"

const pageMyAppointments = "/appointments"
const pageNewAppointment = "/appointment/new"
const pageEditAppointment = "/appointment/edit"

const pageDoctors = "/doctors"

const pageAdminAllAppointments = "/admin/appointments"
const pageAdminSessions = "/admin/sessions"
const pageAdminUsers = "/admin/users"
const pageAdminPaymentEnqueue = "/admin/payment/enqueue"

const pageAdminPaymentDequeue = "/admin/payment/dequeue"
const pageAdminPaymentDequeueToMissedQueue = "/admin/payment/dequeueToMissedQueue"

const pagePaymentQueue = "/payment/queue"
