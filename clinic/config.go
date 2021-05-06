package clinic

import (
	"database/sql"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// True  = clean system with no data
const resetDB = false

// True  = clean system with test data
const resetAndSeedDB = true

// If above 2 are both false = fetch from database

// For doctors' timeslots - 1st consultation @ 8 am, last @ 10 pm
const startOperationHour = 8
const endOperationHour = 22

// 30 mins between each consultations
const appointmentIntervals = 30

// Maximum number of days in the future allowed to make an appointment.
const MaxAdvanceApptDays = 90

// Password policy.
const MinPasswordLength = 8

// Disabled for ease of testing of assignment; Set to true to check for true NRIC format (PDPA though...)
const strictNRIC = false

// DB settings
var (
	db_hostname   string
	db_port       string
	db_username   string
	db_password   string
	db_database   string
	db_connection string
)

// Package globals
var Wg sync.WaitGroup
var mutex sync.Mutex
var clinicDb *sql.DB

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

// Returns database connection string.
func DbConnection() string {
	return db_connection
}

// Sets the singleton database connection to be used by package.
func SetDb(myDb *sql.DB) {
	clinicDb = myDb
}
