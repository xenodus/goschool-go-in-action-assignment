package clinic

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Seed or fetch from DB
const seedDB = false

// For doctors' timeslots - 1st consultation @ 8 am, last @ 10 pm
const startOperationHour = 8
const endOperationHour = 22
const appointmentIntervals = 30 // 30 mins between each consultations

// Password policy
const MinPasswordLength = 8

// Disabled for ease of testing of assignment
const strictNRIC = false

// Set to true if current time 10pm and want to test
// Set current hour minute for testing
const testFakeTime = false
const testHour = 9
const testMinute = 15

// DB
var (
	db_hostname   string
	db_port       string
	db_username   string
	db_password   string
	db_database   string
	db_connection string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		db_hostname = os.Getenv("MYSQL_HOSTNAME")
		db_port = os.Getenv("MYSQL_PORT")
		db_username = os.Getenv("MYSQL_USERNAME")
		db_password = os.Getenv("MYSQL_PASSWORD")
		db_database = os.Getenv("MYSQL_DATABASE")

		db_connection = db_username + ":" + db_password + "@tcp(" + db_hostname + ":" + db_port + ")/" + db_database
	}
}
