package main

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

func seedAdmins() {
	admins = append(admins, []string{"S1234567A", "S0000000B", "S9999999C"}...)
}

func seedPatients() {
	createPatient("S8621568C", "12345678", "Bruce", "Wayne")
	createPatient("S1234567A", "12345678", "Clark", "Kent")   // admin
	createPatient("S0000000B", "12345678", "Diana", "Prince") // admin
}

func seedDoctors() {
	addDoctor("Alvin", "Yeoh")
	addDoctor("Ben", "Low")
	addDoctor("Dina", "Malyana")
	addDoctor("Lydia", "Ng")
	addDoctor("Max", "Wu")
	addDoctor("June", "Yeoh")
	addDoctor("Yonghao", "Fu")
	addDoctor("Geraldine", "Tee")
	addDoctor("Ray", "Chong")
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
