package clinic

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdmins() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		adminIdsString := os.Getenv("ADMIN_IDS")
		adminIds := strings.Split(adminIdsString, ", ")
		Admins = append(Admins, adminIds...)
	}
}

func SeedPatients() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		testAcctPasswordString := os.Getenv("DEFAULT_TEST_PASSWORD")
		bPassword, err := bcrypt.GenerateFromPassword([]byte(testAcctPasswordString), bcrypt.MinCost)

		if err == nil {
			wg.Add(12)
			go CreatePatient("S1111111B", "Anakin", "Skywalker", bPassword)
			go CreatePatient("S2222222C", "Leia", "Organa", bPassword)
			go CreatePatient("S3333333D", "Han", "Solo", bPassword)
			go CreatePatient("S4444444D", "Padmé", "Amidala", bPassword)
			go CreatePatient("S5555555E", "Owen", "Lars", bPassword)
			go CreatePatient("S6666666F", "Qui-Gon", "Jin", bPassword)
			go CreatePatient("S7777777G", "Kanan", "Jarrus", bPassword)
			// admins
			go CreatePatient("S0000000A", "Cal", "Kestis", bPassword)
			go CreatePatient("S1234567A", "Mace", "Windu", bPassword)
			go CreatePatient("S7654321A", "Savage", "Opress", bPassword)
			go CreatePatient("S8888888A", "Orson", "Krennic", bPassword)
			go CreatePatient("S9999999A", "Sheev", "Palpatine", bPassword)
			wg.Wait()
		}
	}
}

func SeedDoctors() {
	wg.Add(10)
	go addDoctor("Boba", "Fett")
	go addDoctor("Bo-Katan", "Kryze")
	go addDoctor("Paz", "Vizsla")
	go addDoctor("Sabine", "Wren")
	go addDoctor("Lando", "Calrissian")
	go addDoctor("Wedge", "Antilles")
	go addDoctor("Cassian", "Andor")
	go addDoctor("Chirrut", "Îmwe")
	go addDoctor("Galen", "Erso")
	go addDoctor("Saw", "Gerrera")
	wg.Wait()
}

func SeedAppointments() {
	no2seed := 10
	rand.Seed(time.Now().Unix())

	for no2seed > 0 {
		randomPat := Patients[rand.Intn(len(Patients))]
		randomDoc := Doctors[rand.Intn(len(Doctors))]

		timeAvailable := GetAvailableTimeslot(append(randomPat.Appointments, randomDoc.Appointments...))

		if len(timeAvailable) > 0 {
			randomTime := timeAvailable[rand.Intn(len(timeAvailable))]
			MakeAppointment(randomTime, randomPat, randomDoc)
		} else {
			fmt.Println("Seeding Appointment Error: No more timeslot for", randomPat.First_name, randomPat.Last_name, "by Dr.", randomDoc.First_name, randomDoc.Last_name)
		}

		no2seed--
	}
}

func SeedPaymentQueue() {
	no2queue := 3
	no2MissedQueue := 0
	rand.Seed(time.Now().Unix())

	for no2queue > 0 {

		if len(Appointments) > 0 {
			appt := Appointments[rand.Intn(len(Appointments))]
			CreatePayment(appt, 29.99)
		}

		no2queue--
	}

	for no2MissedQueue > 0 {

		if PaymentQ.Front != nil {
			PaymentQ.DequeueToMissedPaymentQueue()
		}

		no2MissedQueue--
	}
}
