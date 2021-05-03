package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"assignment4/session"
)

// Globals
var appointments = []*appointment{}
var appointmentsSortedByTimeslot = []*appointment{}
var appointment_start_id int64 = 1000

type appointment struct {
	Id      int64 // unique identifier
	Time    int64 // unix time for easy sorting via int value comparison
	Patient *patient
	Doctor  *doctor
}

// Make and sort by appointment time
func makeAppointment(t int64, pat *patient, doc *doctor) (*appointment, error) {

	mutex.Lock()
	defer mutex.Unlock()

	app := appointment{}
	_, err := isThereTimeslot(pat, doc)

	if err == nil {

		atomic.AddInt64(&appointment_start_id, 1)
		app := appointment{appointment_start_id, t, pat, doc}

		wg.Add(3)
		go addAppointment(&app)
		go app.Doctor.addAppointment(&app)
		go app.Patient.addAppointment(&app)
		wg.Wait()

		return &app, nil
	}

	return &app, err
}

func addAppointment(appt *appointment) {
	defer wg.Done()
	appointments = append(appointments, appt)
	updateTimeslotSortedAppts()
}

func (appt *appointment) editAppointment(t int64, pat *patient, doc *doctor) error {

	mutex.Lock()
	{
		// Update
		appt.Patient = pat
		appt.Doctor = doc
		appt.Time = t

		// Re-sort appointmentsSortedByTimeslot by time
		updateTimeslotSortedAppts()
		// Re-sort doc and patient's appts
		pat.sortAppointments()
		doc.sortAppointments()
	}
	mutex.Unlock()

	return nil
}

func (appt *appointment) cancelAppointment() {
	mutex.Lock()
	{
		apptIDIndex := binarySearchApptID(appt.Id)

		if apptIDIndex >= 0 {
			// Remove from Patient & Doctor
			wg.Add(2)
			go appointments[apptIDIndex].Patient.cancelAppointment(appt.Id)
			go appointments[apptIDIndex].Doctor.cancelAppointment(appt.Id)
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
		}
	}
	mutex.Unlock()
}

// Check if time of appointment is in the past - e.g. process started at 3:55 PM, user chose 4 PM timeslot but submitted at 4:05 PM
func isApptTimeValid(t int64) (bool, error) {

	if !testFakeTime {

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
	mergeSortByTime(tempAppts, 0, len(tempAppts)-1)
	appointmentsSortedByTimeslot = tempAppts
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

func mergeSortByTime(arr []*appointment, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2          // index of midpoint
		mergeSortByTime(arr, first, mid)   // sort left half
		mergeSortByTime(arr, mid+1, last)  // sort right half
		mergeByTime(arr, first, mid, last) // merge the two halves
	}
}

func mergeByTime(arr []*appointment, first int, mid int, last int) {

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
func binarySearchApptID(apptID int64) int {
	return binarySearchAppt(appointments, 0, len(appointments)-1, apptID)
}

func binarySearchAppt(arr []*appointment, first int, last int, apptID int64) int {
	if first > last { // item not found
		return -1
	} else {
		mid := (first + last) / 2

		if arr[mid].Id == apptID { // item found
			return mid
		} else {
			if apptID < arr[mid].Id { // item in first half
				return binarySearchAppt(arr, first, mid-1, apptID) // search in first half
			} else { // item in second half
				return binarySearchAppt(arr, mid+1, last, apptID) // search in second half
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

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Get querystring values
	inputApptId := req.FormValue("apptId")
	action := req.FormValue("action")

	if action != "edit" && action != "cancel" {
		Error.Println(req.RemoteAddr, " Appointment update failure: invalid action type")
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle          string
		User               *patient
		Appt               *appointment
		TimeslotsAvailable []int64
		ErrorMsg           string
		SuccessMsg         string
	}{
		"Edit Appointment", thePatient, nil, nil, "", "",
	}

	apptId, err := strconv.ParseInt(inputApptId, 10, 64)

	if err != nil {
		session.SetNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		Error.Println(req.RemoteAddr, " Appointment update failure: invalid appt id. Unable to parse.")
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Check if appt id is valid
	patientApptIDIndex := binarySearchApptID(apptId)

	if patientApptIDIndex < 0 {
		session.SetNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		Error.Println(req.RemoteAddr, " Appointment update failure:", ErrAppointmentIDNotFound.Error())
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	payload.Appt = appointments[patientApptIDIndex]

	// Does not belong to logged in user
	if payload.Appt.Patient != thePatient {
		session.SetNotification(req, ErrAppointmentIDNotFound.Error(), "Error")
		Error.Println(req.RemoteAddr, " Appointment update failure: appt ", payload.Appt.Id, " does not belong to user ", thePatient.Id)
		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Cancel Appt
	if action == "cancel" {
		if req.Method == http.MethodPost {
			session.SetNotification(req, "Appointment cancelled!", "Success")
			Info.Println(req.RemoteAddr, " Appointment cancelled successfully:", payload.Appt.Id)
			payload.Appt.cancelAppointment()
		} else {
			Error.Println(req.RemoteAddr, " Appointment cancellation failure: GET REQUEST")
		}

		http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
		return
	}

	// Edit Appt
	if action == "edit" {

		payload.TimeslotsAvailable = getAvailableTimeslot(append(payload.Appt.Doctor.Appointments, payload.User.Appointments...))
		_, timeSlotErr := isThereTimeslot(payload.User, payload.Appt.Doctor)

		if timeSlotErr != nil {
			payload.ErrorMsg = ErrNoMoreTimeslot.Error()
			Error.Println(req.RemoteAddr, " Appointment update failure:", payload.ErrorMsg)
			tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
			return
		}

		if req.Method == http.MethodPost {

			// Form submit values
			timeslot := req.FormValue("timeslot")

			if timeslot != "" {

				t, _ := strconv.ParseInt(timeslot, 10, 64)

				// Patient / Doctor time check
				if !payload.Appt.Patient.isFreeAt(t) || !payload.Appt.Doctor.isFreeAt(t) {
					payload.ErrorMsg = ErrDuplicateTimeslot.Error()
					Error.Println(req.RemoteAddr, " Appointment update failure:", payload.ErrorMsg)
					tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
					return
				}

				_, isApptTimeValidErr := isApptTimeValid(t)

				// Past time
				if isApptTimeValidErr != nil {
					payload.ErrorMsg = isApptTimeValidErr.Error()
					Error.Println(req.RemoteAddr, " Appointment update failure:", payload.ErrorMsg)
					tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
					return
				}

				payload.Appt.editAppointment(t, payload.Appt.Patient, payload.Appt.Doctor)
				session.SetNotification(req, "Appointment updated!", "Success")
				Info.Println(req.RemoteAddr, " Appointment updated successfully:", payload.Appt.Id)
				http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "editAppointment.gohtml", payload)
}

func appointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle  string
		User       *patient
		SuccessMsg string
		ErrorMsg   string
	}{
		"My Appointments", thePatient, "", "",
	}

	// Get notifications from session
	if notify, notifyErr := session.GetNotification(req); notifyErr == nil {
		if notify != nil {
			if notify.Type == "Success" {
				payload.SuccessMsg = notify.Message
			} else if notify.Type == "Error" {
				payload.ErrorMsg = notify.Message
			}
			session.ClearNotification(req)
		}
	}

	tpl.ExecuteTemplate(res, "appointments.gohtml", payload)
}

func newAppointmentPage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := isLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Form submit values
	doctorID := req.FormValue("doctor")
	timeslot := req.FormValue("timeslot")

	payload := struct {
		PageTitle          string
		User               *patient
		Doctors            []*doctor
		ChosenDoctor       *doctor
		TimeslotsAvailable []int64
		ErrorMsg           string
	}{
		"New Appointment", thePatient, doctors, nil, nil, "",
	}

	if req.Method == http.MethodPost {

		if doctorID != "" {
			doctorID, _ := strconv.ParseInt(doctorID, 10, 64)
			doc, err := doctorsBST.getDoctorByIDBST(doctorID)

			if err != nil {
				payload.ErrorMsg = err.Error()
				Error.Println(req.RemoteAddr, " Appointment creation failure:", payload.ErrorMsg)
				tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
				return
			}

			payload.ChosenDoctor = doc
			payload.TimeslotsAvailable = getAvailableTimeslot(append(payload.ChosenDoctor.Appointments, thePatient.Appointments...))
			_, timeSlotErr := isThereTimeslot(thePatient, payload.ChosenDoctor)

			if timeSlotErr != nil {
				if timeSlotErr == ErrDoctorNoMoreTimeslot {
					payload.ErrorMsg = "Dr. " + payload.ChosenDoctor.First_name + " " + payload.ChosenDoctor.Last_name + " has no more available timeslots for today"
				} else if timeSlotErr == ErrPatientNoMoreTimeslot {
					payload.ErrorMsg = "You have no more available timeslots for today"
				} else {
					payload.ErrorMsg = timeSlotErr.Error()
				}

				payload.ChosenDoctor = nil

				Error.Println(req.RemoteAddr, " Appointment creation failure:", payload.ErrorMsg)
				tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
				return
			}
		}

		if timeslot != "" && payload.ChosenDoctor != nil && payload.ErrorMsg == "" {
			t, _ := strconv.ParseInt(timeslot, 10, 64)

			// Check if slot truely exists
			if !payload.ChosenDoctor.isFreeAt(t) || !thePatient.isFreeAt(t) {
				payload.ErrorMsg = ErrDuplicateTimeslot.Error()
				Error.Println(req.RemoteAddr, " Appointment creation failure:", payload.ErrorMsg)
				tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
				return
			}

			_, isApptTimeValidErr := isApptTimeValid(t)

			// Past time
			if isApptTimeValidErr != nil {
				payload.ErrorMsg = isApptTimeValidErr.Error()
				Error.Println(req.RemoteAddr, " Appointment creation failure:", payload.ErrorMsg)
				tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
				return
			}

			newAppt, newApptErr := makeAppointment(t, thePatient, payload.ChosenDoctor)

			if newApptErr == nil {
				session.SetNotification(req, "Appointment scheduled!", "Success")
				Info.Println(req.RemoteAddr, " Appointment created successfully:", newAppt.Id)
				http.Redirect(res, req, pageMyAppointments, http.StatusSeeOther)
				return
			}
		}
	}

	tpl.ExecuteTemplate(res, "newAppointment.gohtml", payload)
}
