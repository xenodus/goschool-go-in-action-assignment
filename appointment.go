package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

var appointment_start_id int64 = 1000

type appointment struct {
	Id      int64 // unique identifier
	Time    int64 // unix time for easy sorting via int value comparison
	Patient *patient
	Doctor  *doctor
}

func (appt *appointment) editAppointment(t int64, pat *patient, doc *doctor) error {

	// Update
	appt.Patient = pat
	appt.Doctor = doc
	appt.Time = t

	// Re-sort appointmentsSortedByTimeslot by time
	updateTimeslotSortedAppts()
	// Re-sort doc and patient's appts
	pat.sortAppointments()
	doc.sortAppointments()

	return nil
}

func isApptTimeValid(t int64) (bool, error) {

	if !testFakeTime {

		// Check if time of appointment is in the past - e.g. process started at 3:55PM, use chose 4PM timeslot but submitted at 4:05PM
		currTime := time.Now()
		var lastPossibleTimeslot int64

		if currTime.Minute() >= 30 {
			lastPossibleTimeslot = time.Date(currTime.Year(), currTime.Month(), currTime.Day(), currTime.Hour(), 30, 0, 0, time.Local).Unix()
		} else {
			lastPossibleTimeslot = time.Date(currTime.Year(), currTime.Month(), currTime.Day(), currTime.Hour(), 0, 0, 0, time.Local).Unix()
		}

		if lastPossibleTimeslot > t {
			return false, ErrTimeslotExpired
		}
	}

	return true, nil
}

// Make and sort by appointment time
func makeAppointment(t int64, pat *patient, doc *doctor) (appointment, error) {

	app := appointment{}
	_, err := isThereTimeslot(pat, doc)

	if err == nil {

		atomic.AddInt64(&appointment_start_id, 1)
		app = appointment{appointment_start_id, t, pat, doc}

		appointments = append(appointments, &app) // add to global appointments
		app.Doctor.addAppointment(&app)
		app.Patient.addAppointment(&app)

		updateTimeslotSortedAppts()

		return app, nil
	}

	return app, err
}

func isThereTimeslot(pat *patient, doc *doctor) (bool, error) {
	patientTimeslotsAvailable := getAvailableTimeslot(pat.Appointments)

	if len(patientTimeslotsAvailable) <= 0 {
		return false, ErrPatientNoMoreTimeslot
	}

	doctorTimeslotsAvailable := getAvailableTimeslot(doc.Appointments)

	if len(doctorTimeslotsAvailable) <= 0 {
		return false, ErrDoctorNoMoreTimeslot
	}

	timeslotsAvailable := getAvailableTimeslot(append(doc.Appointments, pat.Appointments...))

	if len(timeslotsAvailable) <= 0 {
		return false, ErrNoMoreTimeslot
	}

	return true, nil
}

func updateTimeslotSortedAppts() {
	tempAppts := make([]*appointment, len(appointments))
	copy(tempAppts, appointments)
	mergeSort(tempAppts, 0, len(tempAppts)-1)
	appointmentsSortedByTimeslot = tempAppts
}

func cancelAppointment(apptID int64) error {

	apptIDIndex := binarySearchApptID(appointments, 0, len(appointments)-1, apptID)

	if apptIDIndex >= 0 {

		// Remove from Patient & Doctor
		// Concurrently handle removal of pointer from the individual slices
		wg.Add(2)
		go appointments[apptIDIndex].Patient.cancelAppointment(apptID)
		go appointments[apptIDIndex].Doctor.cancelAppointment(apptID)
		wg.Wait()

		if apptIDIndex == 0 {
			appointments = appointments[1:]
		} else if apptIDIndex == len(appointments)-1 {
			appointments = appointments[:apptIDIndex]
		} else {
			appointments = append(appointments[:apptIDIndex], appointments[apptIDIndex+1:]...)
		}

		// Re-sort appointmentsSortedByTimeslot by time
		updateTimeslotSortedAppts()
		return nil
	}

	return ErrAppointmentIDNotFound
}

func getAvailableTimeslot(apptsToExclude []*appointment) []int64 {

	allTimeSlots := timeSlotsGenerator()
	availableTimeslots := []int64{}

	// Exclude timeslots that are already occipied by patient and doctor
	for _, v := range allTimeSlots {
		var exists = false

		for _, v2 := range apptsToExclude {
			if v == v2.Time {
				exists = true
			}
		}

		if !exists {
			availableTimeslots = append(availableTimeslots, v)
		}
	}

	return availableTimeslots
}

// Returns slice of available time slots in 30 mins intervals from current time in unix/epoch time
func timeSlotsGenerator() []int64 {
	currentTimeHour := time.Now().Hour()
	currentTimeMinute := time.Now().Minute()

	// Test overwrites
	if testFakeTime {
		currentTimeHour = testHour
		currentTimeMinute = testMinute
	}

	timeSlots := []int64{}
	startHour := currentTimeHour
	startMinute := 30
	currTime := time.Now()

	if currentTimeHour < startOperationHour || currentTimeHour > endOperationHour {
		startHour = startOperationHour
		startMinute = 0

		if currentTimeHour > endOperationHour {
			currTime = currTime.Add(time.Hour * 24) // next day
		}
	} else {
		if currentTimeMinute >= appointmentIntervals {
			startHour += 1
			startMinute = 0
		}
	}

	for startHour <= endOperationHour {

		t := time.Date(currTime.Year(), currTime.Month(), currTime.Day(), startHour, startMinute, 0, 0, time.Local)
		timeSlots = append(timeSlots, t.Unix())

		// Next iteration
		if startMinute == 0 {
			if startHour == endOperationHour {
				break
			}
			startMinute = 30
		} else {
			startMinute = 0
		}

		if startMinute == 0 {
			startHour += 1
		}
	}

	return timeSlots
}

func mergeSort(arr []*appointment, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2    // index of midpoint
		mergeSort(arr, first, mid)   // sort left half
		mergeSort(arr, mid+1, last)  // sort right half
		merge(arr, first, mid, last) // merge the two halves
	}
}

func merge(arr []*appointment, first int, mid int, last int) {

	tempArr := make([]*appointment, len(arr))

	// initialize the local indexes to indicate the subarrays
	first1 := first   // beginning of first subarray
	last1 := mid      // end of first subarray
	first2 := mid + 1 // beginning of second subarray
	last2 := last     // end of second subarray

	// while both subarrays are not empty, copy the
	// smaller item into the temporary array
	index := first1 // next available location in tempArray
	for (first1 <= last1) && (first2 <= last2) {
		if arr[first1].Time < arr[first2].Time {
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

// Binary search for appointment id in sorted slice
func binarySearchApptID(arr []*appointment, first int, last int, apptID int64) int {
	if first > last { // item not found
		return -1
	} else {
		mid := (first + last) / 2
		if arr[mid].Id == apptID { // item found
			return mid
		} else {
			if apptID < arr[mid].Id { // item in first half
				return binarySearchApptID(arr, first, mid-1, apptID) // search in first half
			} else { // item in second half
				return binarySearchApptID(arr, mid+1, last, apptID) // search in second half
			}
		}
	}
}

// Sequential search
func searchApptID(arr []*appointment, apptID int64) (int, error) {
	for k, v := range arr {
		if v.Id == apptID {
			return k, nil
		}
	}
	return -1, ErrAppointmentIDNotFound
}

// Web Pages

func editAppointmentPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	// Get querystring values
	apptId := req.FormValue("apptId")
	action := req.FormValue("action")

	// Form submit values
	timeslot := req.FormValue("timeslot")

	var chosenDoctor *doctor = nil
	var theAppt *appointment = nil
	var timeslotsAvailable []int64
	var errorMsg = ""

	if action == "edit" || action == "cancel" {

		apptId, err := strconv.ParseInt(apptId, 10, 64)

		if err != nil {
			errorMsg = ErrAppointmentIDNotFound.Error()
		} else {
			// Check if appt id is valid
			patientApptIDIndex := binarySearchApptID(thePatient.Appointments, 0, len(thePatient.Appointments)-1, apptId)

			if patientApptIDIndex < 0 {
				errorMsg = ErrAppointmentIDNotFound.Error()
			} else {

				// Cancel Appt
				if action == "cancel" {
					cancelAppointment(apptId)
					http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
					return
				}

				// Edit Appt
				if action == "edit" {

					theAppt = thePatient.Appointments[patientApptIDIndex]
					chosenDoctor = theAppt.Doctor

					timeslotsAvailable = getAvailableTimeslot(append(chosenDoctor.Appointments, thePatient.Appointments...))
					_, timeSlotErr := isThereTimeslot(thePatient, chosenDoctor)

					if timeSlotErr != nil {
						errorMsg = timeSlotErr.Error()
					}

					if timeslot != "" && chosenDoctor != nil {
						t, _ := strconv.ParseInt(timeslot, 10, 64)

						if isApptTimeValid, isApptTimeValidErr := isApptTimeValid(t); isApptTimeValid {
							theAppt.editAppointment(t, thePatient, chosenDoctor)
							http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
							return
						} else {
							errorMsg = isApptTimeValidErr.Error()
						}
					}
				}
			}
		}

		// Anonymous payload
		payload := struct {
			PageTitle          string
			User               *patient
			Doctors            []*doctor
			Appt               *appointment
			ChosenDoctor       *doctor
			TimeslotsAvailable []int64
			ErrorMsg           string
		}{
			"Edit Appointment",
			thePatient,
			doctors,
			theAppt,
			chosenDoctor,
			timeslotsAvailable,
			errorMsg,
		}

		tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
		return
	}

	http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
}

func appointmentPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *patient
	}{
		"My Appointments",
		thePatient,
	}

	tpl.ExecuteTemplate(res, "appointments.gohtml", payload)
}

func newAppointmentPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	thePatient := getLoggedInPatient(res, req)

	// Form submit values
	doctorID := req.FormValue("doctor")
	timeslot := req.FormValue("timeslot")
	errorCode := req.FormValue("error")
	errorTimeslot := req.FormValue("t")

	var chosenDoctor *doctor = nil
	var timeslotsAvailable []int64
	var errorMsg = ""

	if doctorID != "" {
		doctorID, _ := strconv.ParseInt(doctorID, 10, 64)
		doc, err := doctorsBST.getDoctorByIDBST(doctorID)

		if err != nil {
			errorMsg = err.Error()
		} else {
			chosenDoctor = doc
			timeslotsAvailable = getAvailableTimeslot(append(chosenDoctor.Appointments, thePatient.Appointments...))
			_, timeSlotErr := isThereTimeslot(thePatient, chosenDoctor)

			if timeSlotErr != nil {
				errorMsg = timeSlotErr.Error()
			}
		}
	}

	if timeslot != "" && chosenDoctor != nil {
		t, _ := strconv.ParseInt(timeslot, 10, 64)

		// Check if slot truely exists
		if !chosenDoctor.isFreeAt(t) {
			fmt.Println("Doc ain't free")
			http.Redirect(res, req, pageNewAppointment+"?error=dnf&t="+timeslot, http.StatusSeeOther)
			return
		}

		if !thePatient.isFreeAt(t) {
			fmt.Println("Patient ain't free")
			http.Redirect(res, req, pageNewAppointment+"?error=pnf&t"+timeslot, http.StatusSeeOther)
			return
		}

		if isApptTimeValid, isApptTimeValidErr := isApptTimeValid(t); isApptTimeValid {
			_, newApptErr := makeAppointment(t, thePatient, chosenDoctor)

			if newApptErr == nil {
				http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
				return
			}
		} else {
			errorMsg = isApptTimeValidErr.Error()
		}
	}

	if errorCode == "dnf" {
		t, _ := strconv.ParseInt(errorTimeslot, 10, 64)
		errorMsg = "Error: Doctor isn't available at " + time2HumanReadable(t)
	} else if errorCode == "pnf" {
		t, _ := strconv.ParseInt(errorTimeslot, 10, 64)
		errorMsg = "Error: You already another appointment scheduled at " + time2HumanReadable(t)
	}

	// Anonymous payload
	payload := struct {
		PageTitle          string
		User               *patient
		Doctors            []*doctor
		ChosenDoctor       *doctor
		TimeslotsAvailable []int64
		ErrorMsg           string
	}{
		"New Appointment",
		thePatient,
		doctors,
		chosenDoctor,
		timeslotsAvailable,
		errorMsg,
	}

	tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
}
