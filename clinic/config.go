package clinic

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
