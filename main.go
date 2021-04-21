// Author: Alvin Yeoh

package main

import (
	"html/template"
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
var mapSessions = make(map[string]session)

var cookieID string

func init() {
	// Mandatory Test Data
	seedDoctors()
	seedAdmins()

	// Optional Test Data
	seedPatients()
	seedAppointments()
	seedPaymentQueue()

	// Randomizing the cookie name on each init
	cookieID = getRandomCookiePrefix()

	// Adding helper functions to templates
	funcMap := template.FuncMap{
		"time2HumanReadable": time2HumanReadable,
		"getUserByID":        getUserByID,
		"ucFirst":            ucFirst,
		"stripSpace":         stripSpace,
	}

	tpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*"))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	startHttpServer()
}
