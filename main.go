// Author: Alvin Yeoh

package main

import (
	"html/template"
	"sync"
)

// Globals
var wg sync.WaitGroup
var mutex sync.Mutex

var tpl *template.Template

func init() {
	// Mandatory Test Data
	seedDoctors()
	seedAdmins()

	// Optional Test Data
	seedPatients()
	seedAppointments()
	seedPaymentQueue()
}

func main() {
	startHttpServer()
}
