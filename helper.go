package main

import (
	"fmt"
	"math/rand"
	"time"
	"unicode"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func seedAdmins() {
	adminIds := []string{
		"S1234567A",
		"S0000000A",
		"S9999999C",
	}

	admins = append(admins, adminIds...)
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

func seedAppointments() {
	no2seed := 10
	rand.Seed(time.Now().Unix())

	for no2seed > 0 {
		randomPat := patients[rand.Intn(len(patients))]
		randomDoc := doctors[rand.Intn(len(doctors))]

		timeAvailable := getAvailableTimeslot(append(randomPat.Appointments, randomDoc.Appointments...))

		if len(timeAvailable) > 0 {
			randomTime := timeAvailable[rand.Intn(len(timeAvailable))]
			makeAppointment(randomTime, randomPat, randomDoc)
		} else {
			fmt.Println("Seeding Appointment Error: No more timeslot for", randomPat.First_name, randomPat.Last_name, "by Dr.", randomDoc.First_name, randomDoc.Last_name)
		}

		no2seed--
	}
}

func seedPaymentQueue() {
	no2queue := 5
	no2MissedQueue := 2
	rand.Seed(time.Now().Unix())

	for no2queue > 0 {

		if len(appointments) > 0 {
			appt := appointments[rand.Intn(len(appointments))]
			createPayment(appt, 29.99)
		}

		no2queue--
	}

	for no2MissedQueue > 0 {

		if paymentQ.Front != nil {
			paymentQ.dequeueToMissedPaymentQueue()
		}

		no2MissedQueue--
	}
}

func time2HumanReadable(t int64) string {
	return time.Unix(t, 0).Format("3:04PM")
}

func getRandomCookiePrefix() string {
	randomUUIDByte, _ := uuid.NewV4()
	return "CO-" + randomUUIDByte.String()
}

func getUserByID(uid string) *patient {
	user, err := getPatientByID(uid)

	if err == nil {
		return user
	}

	return nil
}

func ucFirst(str string) string {
	if len(str) == 0 {
		return ""
	}
	tmp := []rune(str)
	tmp[0] = unicode.ToUpper(tmp[0])
	return string(tmp)
}
