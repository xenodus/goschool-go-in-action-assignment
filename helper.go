package main

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func seedAdmins() {
	admins = append(admins, []string{"S1234567A", "S0000000A", "S9999999C"}...)
}

func seedPatients() {
	bPassword, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)

	if err == nil {
		wg.Add(7)
		go createPatient("S0000000A", "Diana", "Prince", bPassword) // admin
		go createPatient("S1111111B", "Barry", "Allen", bPassword)
		go createPatient("S2222222C", "Bruce", "Wayne", bPassword)
		go createPatient("S3333333D", "Hal", "Jordan", bPassword)
		go createPatient("S4444444D", "Arthur", "Curry", bPassword)
		go createPatient("S1234567A", "Clark", "Kent", bPassword)   // admin
		go createPatient("S9999999C", "Oliver", "Queen", bPassword) // admin
		wg.Wait()
	}
}

func seedDoctors() {
	wg.Add(9)
	go addDoctor("Ben", "Low")
	go addDoctor("Dina", "Malyana")
	go addDoctor("Lydia", "Ng")
	go addDoctor("Max", "Wu")
	go addDoctor("June", "Yeoh")
	go addDoctor("Yonghao", "Fu")
	go addDoctor("Geraldine", "Tee")
	go addDoctor("Bruce", "Banner")
	go addDoctor("Steven", "Strange")
	wg.Wait()
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
