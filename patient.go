package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type patient struct {
	Id           string
	First_name   string
	Last_name    string
	password     []byte
	Appointments []*appointment
}

func (p *patient) isFreeAt(t int64) bool {
	for _, v := range p.Appointments {
		if v.Time == t {
			return false
		}
	}

	return true
}

func (p *patient) delete() error {
	// 1. remove all appointment from appointments slice with patient in em
	if len(p.Appointments) > 0 {
		for _, v := range p.Appointments {
			cancelAppointment(v.Id)
		}
	}

	// 2. remove sessions with user id
	if len(mapSessions) > 0 {
		for _, v := range mapSessions {
			if v == p.Id {
				delete(mapSessions, v)
			}
		}
	}

	// 3. remove patient from patients slice
	patientIDIndex := binarySearchPatientID(patients, 0, len(patients)-1, p.Id)

	if patientIDIndex >= 0 {

		mutex.Lock()
		{
			if patientIDIndex == 0 {
				patients = patients[1:]
			} else if patientIDIndex == len(patients)-1 {
				patients = patients[:patientIDIndex]
			} else {
				patients = append(patients[:patientIDIndex], patients[patientIDIndex+1:]...)
			}
		}
		mutex.Unlock()
	}

	return nil
}

func getPatientByID(patientID string) (*patient, error) {

	patientIDIndex := binarySearchPatientID(patients, 0, len(patients)-1, patientID)

	if patientIDIndex >= 0 {
		return patients[patientIDIndex], nil
	}

	return nil, ErrPatientIDNotFound
}

func (p *patient) sortAppointments() {
	mergeSort(p.Appointments, 0, len(p.Appointments)-1)
}

func (p *patient) addAppointment(appt *appointment) {
	defer wg.Done()

	var mutex sync.Mutex

	mutex.Lock()
	{
		p.Appointments = append(p.Appointments, appt)
		p.sortAppointments()
	}
	mutex.Unlock()
}

func (p *patient) cancelAppointment(apptID int64) error {
	defer wg.Done()

	apptIDIndex, err := searchApptID(p.Appointments, apptID)

	if apptIDIndex >= 0 {

		mutex.Lock()
		{
			if apptIDIndex == 0 {
				p.Appointments = p.Appointments[1:]
			} else if apptIDIndex == len(p.Appointments)-1 {
				p.Appointments = p.Appointments[:apptIDIndex]
			} else {
				p.Appointments = append(p.Appointments[:apptIDIndex], p.Appointments[apptIDIndex+1:]...)
			}
		}
		mutex.Unlock()

		return nil
	}

	return err
}

// Sorts patients slice alphabetically
func mergeSortPatient(arr []*patient, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2           // index of midpoint
		mergeSortPatient(arr, first, mid)   // sort left half
		mergeSortPatient(arr, mid+1, last)  // sort right half
		mergePatient(arr, first, mid, last) // merge the two halves
	}
}

func mergePatient(arr []*patient, first int, mid int, last int) {

	tempArr := make([]*patient, len(arr))

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
func binarySearchPatientID(arr []*patient, first int, last int, patientID string) int {
	if first > last { // item not found
		return -1
	} else {
		mid := (first + last) / 2
		if arr[mid].Id == patientID { // item found
			return mid
		} else {
			if patientID < arr[mid].Id { // item in first half
				return binarySearchPatientID(arr, first, mid-1, patientID) // search in first half
			} else { // item in second half
				return binarySearchPatientID(arr, mid+1, last, patientID) // search in second half
			}
		}
	}
}

func (p patient) IsAdmin() bool {
	return isAdminCheck(p.Id, 0)
}

// recursion
func isAdminCheck(adminID string, index int) bool {

	if index >= len(admins) {
		return false
	} else {
		if admins[index] == adminID {
			return true
		} else {
			return isAdminCheck(adminID, index+1)
		}
	}
}

// Validate nric
// Translated from https://gist.github.com/kamerk22/ed5e0778b3723311d8dd074c792834ef
func isNRICValid(nric string) bool {

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

func isLoggedIn(req *http.Request) bool {
	myCookie, err := req.Cookie(cookieID)
	if err != nil {
		return false
	}
	username := mapSessions[myCookie.Value]
	_, noPatientErr := getPatientByID(username)

	return noPatientErr == nil
}

func getLoggedInPatient(res http.ResponseWriter, req *http.Request) *patient {
	// get current session cookie
	myCookie, err := req.Cookie(cookieID)
	// not found
	if err != nil {
		id, _ := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:     cookieID,
			Value:    id.String(),
			Path:     pageIndex,
			HttpOnly: true,
		}

		http.SetCookie(res, myCookie)
	}

	// if the patient exists already, get patient
	var thePatient *patient

	if username, ok := mapSessions[myCookie.Value]; ok {
		thePatient, _ = getPatientByID(username)
	}

	return thePatient
}

func createPatient(username, first_name, last_name string, password []byte) *patient {
	defer wg.Done()

	thePatient := patient{username, first_name, last_name, password, nil}
	mutex.Lock()
	{
		patients = append(patients, &thePatient)
		// Sort by patient id alphabetically
		mergeSortPatient(patients, 0, len(patients)-1)
	}
	mutex.Unlock()

	return &thePatient
}

// Web Pages

func registerPage(res http.ResponseWriter, req *http.Request) {

	if isLoggedIn(req) {
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// process form submission
	if req.Method == http.MethodPost {

		// get form values
		username := req.FormValue("nric")
		password := req.FormValue("password")
		firstname := req.FormValue("firstname")
		lastname := req.FormValue("lastname")

		if username != "" {
			if !isNRICValid(username) {
				http.Error(res, "Invalid NRIC", http.StatusForbidden)
				return
			}

			// check if username exist/ taken
			if _, err := getPatientByID(username); err == nil {
				http.Error(res, "You are already registered.", http.StatusForbidden)
				return
			}

			// create session
			id, _ := uuid.NewV4()
			myCookie := &http.Cookie{
				Name:     cookieID,
				Value:    id.String(),
				Path:     pageIndex,
				HttpOnly: true,
			}
			http.SetCookie(res, myCookie)
			mapSessions[myCookie.Value] = username

			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				return
			}

			wg.Add(1)
			createPatient(username, firstname, lastname, bPassword)
			wg.Wait()
		}
		// redirect to main index
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
	}{
		"Register",
		nil,
	}

	tpl.ExecuteTemplate(res, "register.gohtml", payload)
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if isLoggedIn(req) {
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// process form submission
	if req.Method == http.MethodPost {
		username := req.FormValue("nric")
		password := req.FormValue("password")
		// check if user exist with username
		myUser, noPatientErr := getPatientByID(username)

		if noPatientErr != nil {
			http.Error(res, "NRIC and/or password do not match", http.StatusUnauthorized)
			return
		}
		// Matching of password entered
		err := bcrypt.CompareHashAndPassword(myUser.password, []byte(password))
		if err != nil {
			http.Error(res, "NRIC and/or password do not match", http.StatusForbidden)
			return
		}
		// create session
		id, _ := uuid.NewV4()
		myCookie := &http.Cookie{
			Name:     cookieID,
			Value:    id.String(),
			Path:     pageIndex,
			HttpOnly: true,
		}

		http.SetCookie(res, myCookie)
		mapSessions[myCookie.Value] = username
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
	}{
		"Login",
		nil,
	}

	tpl.ExecuteTemplate(res, "login.gohtml", payload)
}

func logoutPage(res http.ResponseWriter, req *http.Request) {
	if !isLoggedIn(req) {
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	req.Cookies()

	myCookie, _ := req.Cookie(cookieID)
	// delete the session
	delete(mapSessions, myCookie.Value)
	// remove the cookie
	expire := time.Now().Add(-7 * 24 * time.Hour)
	myCookie = &http.Cookie{
		Name:     cookieID,
		Value:    "",
		Path:     pageIndex,
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  expire,
	}
	http.SetCookie(res, myCookie)

	http.Redirect(res, req, pageIndex, http.StatusSeeOther)
}

func profilePage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	// Form submit values
	if req.Method == "POST" {
		first_name := req.FormValue("firstname")
		last_name := req.FormValue("lastname")
		password := req.FormValue("password")

		if first_name != "" {
			thePatient.First_name = first_name
		}

		if last_name != "" {
			thePatient.Last_name = last_name
		}

		if password != "" {
			bPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			thePatient.password = bPassword
		}
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
	}{
		"Profile",
		thePatient,
	}

	tpl.ExecuteTemplate(res, "profile.gohtml", payload)
}
