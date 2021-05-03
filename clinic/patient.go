package clinic

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"assignment4/session"
)

// Globals
var Patients = []*Patient{}
var Admins = []string{}

type Patient struct {
	Id           string
	First_name   string
	Last_name    string
	Password     []byte
	Appointments []*Appointment
}

func CreatePatient(username, first_name, last_name string, password []byte) {
	defer wg.Done()

	mutex.Lock()
	{
		thePatient := Patient{username, first_name, last_name, password, nil}
		Patients = append(Patients, &thePatient)
		// Sort by patient id alphabetically
		mergeSortPatient(Patients, 0, len(Patients)-1)
	}
	mutex.Unlock()
}

func (p *Patient) EditPatient(username, first_name, last_name string, password []byte) {
	mutex.Lock()
	{
		p.Id = username
		p.First_name = first_name
		p.Last_name = last_name
		p.Password = password
		// Sort by patient id alphabetically
		mergeSortPatient(Patients, 0, len(Patients)-1)
	}
	mutex.Unlock()
}

func (p *Patient) IsFreeAt(t int64) bool {
	for _, v := range p.Appointments {
		if v.Time == t {
			return false
		}
	}

	return true
}

func (p *Patient) DeletePatient() error {

	// 1. remove all appointment from appointments slice with patient in em
	for len(p.Appointments) > 0 {
		p.Appointments[0].CancelAppointment()
	}

	// 2. remove sessions with user id
	mutex.Lock()
	{
		if len(session.MapSessions) > 0 {
			for k, v := range session.MapSessions {
				if v.Id == p.Id {
					delete(session.MapSessions, k)
				}
			}
		}

		// 3. remove patient from patients slice
		patientIDIndex := binarySearchPatientID(p.Id)

		if patientIDIndex >= 0 {

			if patientIDIndex == 0 {
				Patients = Patients[1:]
			} else if patientIDIndex == len(Patients)-1 {
				Patients = Patients[:patientIDIndex]
			} else {
				Patients = append(Patients[:patientIDIndex], Patients[patientIDIndex+1:]...)
			}
		}
	}
	mutex.Unlock()

	return nil
}

func GetPatientByID(patientID string) (*Patient, error) {

	patientIDIndex := binarySearchPatientID(patientID)

	if patientIDIndex >= 0 {
		return Patients[patientIDIndex], nil
	}

	return nil, ErrPatientIDNotFound
}

func (p *Patient) sortAppointments() {
	mergeSortByTime(p.Appointments, 0, len(p.Appointments)-1) // sorted by time
}

func (p *Patient) addAppointment(appt *Appointment) {
	defer wg.Done()
	p.Appointments = append(p.Appointments, appt)
	p.sortAppointments()
}

func (p *Patient) cancelAppointment(apptID int64) error {
	defer wg.Done()

	apptIDIndex, err := searchApptID(p.Appointments, apptID)

	if apptIDIndex >= 0 {

		if apptIDIndex == 0 {
			p.Appointments = p.Appointments[1:]
		} else if apptIDIndex == len(p.Appointments)-1 {
			p.Appointments = p.Appointments[:apptIDIndex]
		} else {
			p.Appointments = append(p.Appointments[:apptIDIndex], p.Appointments[apptIDIndex+1:]...)
		}

		return nil
	}

	return err
}

// Sorts patients slice alphabetically
func mergeSortPatient(arr []*Patient, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2           // index of midpoint
		mergeSortPatient(arr, first, mid)   // sort left half
		mergeSortPatient(arr, mid+1, last)  // sort right half
		mergePatient(arr, first, mid, last) // merge the two halves
	}
}

func mergePatient(arr []*Patient, first int, mid int, last int) {

	tempArr := make([]*Patient, len(arr))

	// initialize the local indexes to indicate the subarrays
	first1 := first   // beginning of first subarray
	last1 := mid      // end of first subarray
	first2 := mid + 1 // beginning of second subarray
	last2 := last     // end of second subarray

	// while both subarrays are not empty, copy the
	// smaller item into the temporary array
	index := first1 // next available location in tempArray
	for (first1 <= last1) && (first2 <= last2) {
		if arr[first1].Id < arr[first2].Id {
			tempArr[index] = arr[first1]
			first1++
		} else {
			tempArr[index] = arr[first2]
			first2++
		}

		index++
	}

	// finish off the nonempty subarray
	// finish off the first subarray, if necessary
	for first1 <= last1 {
		tempArr[index] = arr[first1]
		first1++
		index++
	}

	// finish off the second subarray, if necessary
	for first2 <= last2 {
		tempArr[index] = arr[first2]
		first2++
		index++
	}

	// copy the result back into the original array
	for index = first; index <= last; index++ {
		arr[index] = tempArr[index]
	}
}

// Binary search for patient id in sorted slice
func binarySearchPatientID(patientID string) int {
	return binarySearchPatient(Patients, 0, len(Patients)-1, patientID)
}

func binarySearchPatient(arr []*Patient, first int, last int, patientID string) int {
	if first > last { // item not found
		return -1
	} else {
		mid := (first + last) / 2
		if arr[mid].Id == patientID { // item found
			return mid
		} else {
			if patientID < arr[mid].Id { // item in first half
				return binarySearchPatient(arr, first, mid-1, patientID) // search in first half
			} else { // item in second half
				return binarySearchPatient(arr, mid+1, last, patientID) // search in second half
			}
		}
	}
}

func (p Patient) IsAdmin() bool {
	return isAdminCheck(p.Id, 0)
}

// Recursion
func isAdminCheck(adminID string, index int) bool {

	if index >= len(Admins) {
		return false
	} else {
		if Admins[index] == adminID {
			return true
		} else {
			return isAdminCheck(adminID, index+1)
		}
	}
}

// Validate NRIC
// Translated from https://gist.github.com/kamerk22/ed5e0778b3723311d8dd074c792834ef
func IsNRICValid(nric string) bool {

	if strictNRIC == false {
		return len(nric) == 9
	} else {

		if len(nric) != 9 {
			return false
		}

		icNoArr := [7]int{}
		nric = strings.ToUpper(nric)
		nricRunes := []rune(nric)

		icNoArr[0], _ = strconv.Atoi(string(nricRunes[1]))
		icNoArr[1], _ = strconv.Atoi(string(nricRunes[2]))
		icNoArr[2], _ = strconv.Atoi(string(nricRunes[3]))
		icNoArr[3], _ = strconv.Atoi(string(nricRunes[4]))
		icNoArr[4], _ = strconv.Atoi(string(nricRunes[5]))
		icNoArr[5], _ = strconv.Atoi(string(nricRunes[6]))
		icNoArr[6], _ = strconv.Atoi(string(nricRunes[7]))

		icNoArr[0] *= 2
		icNoArr[1] *= 7
		icNoArr[2] *= 6
		icNoArr[3] *= 5
		icNoArr[4] *= 4
		icNoArr[5] *= 3
		icNoArr[6] *= 2

		weight := 0

		for _, v := range icNoArr {
			weight += v
		}

		offset := 0

		if string(nricRunes[0]) == "T" || string(nricRunes[0]) == "G" {
			offset = 4
		}

		tmp := math.Mod(float64(offset+weight), 11)

		st := [11]string{"J", "Z", "I", "H", "G", "F", "E", "D", "C", "B", "A"}
		fg := [11]string{"X", "W", "U", "T", "R", "Q", "P", "N", "M", "L", "K"}

		theAlpha := ""

		if string(nricRunes[0]) == "S" || string(nricRunes[0]) == "T" {
			theAlpha = st[int(tmp)]
		} else if string(nricRunes[0]) == "F" || string(nricRunes[0]) == "G" {
			theAlpha = fg[int(tmp)]
		}

		return string(nricRunes[8]) == theAlpha
	}
}

func IsLoggedIn(req *http.Request) (*Patient, bool) {
	myCookie, err := req.Cookie(session.CookieID)
	if err != nil {
		return nil, false
	}

	username := session.MapSessions[myCookie.Value].Id
	patient, noPatientErr := GetPatientByID(username)

	if noPatientErr == nil {
		// also update session with last access datetime
		newSession := session.MapSessions[myCookie.Value]
		newSession.LastModified = time.Now().Unix()
		newSession.LastVisited = req.URL
		session.MapSessions[myCookie.Value] = newSession
	}

	return patient, noPatientErr == nil
}
