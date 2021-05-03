package main

import (
	"html/template"
	"io"
	"log"
	"os"
	"sync"
)

// Globals
var wg sync.WaitGroup
var mutex sync.Mutex
var tpl *template.Template

// Logs
var Trace *log.Logger   // Just about anything
var Info *log.Logger    // Important information
var Warning *log.Logger // Be concerned
var Error *log.Logger   // Critical problem

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

	file, err := os.OpenFile("./logs/out.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	defer file.Close()

	// Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(io.MultiWriter(os.Stdout, file), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(os.Stderr, file), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, file), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	startHttpServer()
}
