package main

import (
	"assignment4/clinic"
	"assignment4/web"
)

func init() {
	// Mandatory Test Data
	clinic.SeedDoctors()
	clinic.SeedAdmins()

	// Optional Test Data
	clinic.SeedPatients()
	clinic.SeedAppointments()
	clinic.SeedPaymentQueue()
}

func main() {
	web.StartHttpServer()
}
