package main

import (
	"assignment4/clinic"
	"assignment4/web"
)

func init() {
	// Truncate DB and Seed Test Data
	clinic.SeedData()
}

func main() {
	web.StartHttpServer()
}
