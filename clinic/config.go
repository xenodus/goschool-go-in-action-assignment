package clinic

import (
	"database/sql"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// True  = reset system.
const resetDB = false

// True  = reset system and seed with test data.
const resetAndSeedDB = false

// If above 2 are both false = fetch from database.

// Doctors' timeslots - First consultation at 8 am.
const startOperationHour = 8

// Doctors' timeslots - Last consultation at 10 pm.
const endOperationHour = 22

// 30 mins between each consultations.
const appointmentIntervals = 30

// Maximum number of days in the future allowed to make an appointment.
const MaxAdvanceApptDays = 90

// Password policy.
const MinPasswordLength = 8

// Disabled for ease of testing of assignment; Set to true to check for true NRIC format.
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

// DbConnection returns the database connection string.
func DbConnection() string {
	return db_connection
}

// SetDb sets the singleton database connection to be used by package.
func SetDb(myDb *sql.DB) {
	clinicDb = myDb
}
