package main

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func seedAdmins() {
	admins = append(admins, []string{"S1234567A", "S0000000B", "S9999999C"}...)
}

func seedPatients() {
	bPassword, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)

	if err == nil {
		go createPatient("S8621568C", "Bruce", "Wayne", bPassword)
		go createPatient("S1234567A", "Clark", "Kent", bPassword)   // admin
		go createPatient("S0000000B", "Diana", "Prince", bPassword) // admin
	}
}

func seedDoctors() {
	go addDoctor("Alvin", "Yeoh")
	go addDoctor("Ben", "Low")
	go addDoctor("Dina", "Malyana")
	go addDoctor("Lydia", "Ng")
	go addDoctor("Max", "Wu")
	go addDoctor("June", "Yeoh")
	go addDoctor("Yonghao", "Fu")
	go addDoctor("Geraldine", "Tee")
	go addDoctor("Ray", "Chong")
}

func time2HumanReadable(t int64) string {
	return time.Unix(t, 0).Format("3:04PM")
}

func getRandomCookiePrefix() string {
	randomUUIDByte, _ := uuid.NewV4()
	return "CO-" + randomUUIDByte.String()
}

func isUserAdminByID(uid string) bool {
	p, err := getPatientByID(uid)

	if err == nil {
		return p.IsAdmin()
	}

	return false
}
