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

var paymentQ = paymentQueue{}
var missedPaymentQ = paymentQueue{}
var appointmentsSortedByTimeslot = []*appointment{}

var admins = []string{}

var doctorsBST *BST
var wg sync.WaitGroup
var mutex sync.Mutex

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
