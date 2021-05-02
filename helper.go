package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func seedAdmins() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		adminIdsString := os.Getenv("ADMIN_IDS")
		adminIds := strings.Split(adminIdsString, ",")
		admins = append(admins, adminIds...)
	}
}

func seedPatients() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		testAcctPasswordString := os.Getenv("DEFAULT_TEST_PASSWORD")
		bPassword, err := bcrypt.GenerateFromPassword([]byte(testAcctPasswordString), bcrypt.MinCost)

		if err == nil {
			wg.Add(10)
			go createPatient("S1111111B", "Barry", "Allen", bPassword)
			go createPatient("S2222222C", "Bruce", "Wayne", bPassword)
			go createPatient("S3333333D", "Hal", "Jordan", bPassword)
			go createPatient("S4444444D", "Arthur", "Curry", bPassword)
			go createPatient("S5555555E", "Jay", "Garrick", bPassword)
			go createPatient("S6666666F", "John", "Steward", bPassword)
			go createPatient("S7777777G", "Wally", "West", bPassword)
			// admins
			go createPatient("S0000000A", "Diana", "Prince", bPassword)
			go createPatient("S1234567A", "Clark", "Kent", bPassword)
			go createPatient("S9999999A", "Oliver", "Queen", bPassword)
			wg.Wait()
		}
	}
}

func seedDoctors() {
	wg.Add(10)
	go addDoctor("Steve", "Rogers")
	go addDoctor("Tony", "Stark")
	go addDoctor("Peter", "Parker")
	go addDoctor("Sam", "Wilson")
	go addDoctor("Clint", "Barton")
	go addDoctor("Wanda", "Maximoff")
	go addDoctor("Scott", "Lang")
	go addDoctor("Bruce", "Banner")
	go addDoctor("Steven", "Strange")
	go addDoctor("Carol", "Denvers")
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
	no2MissedQueue := 0
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

func stripSpace(str string) string {
	return strings.ReplaceAll(str, " ", "")
}
