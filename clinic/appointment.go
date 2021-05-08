package clinic

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Appointments holds all the appointments sorted by Id.
var Appointments = []*Appointment{}

// AppointmentsSortedByTimeslot holds all appointments sorted by time.
var AppointmentsSortedByTimeslot = []*Appointment{}

type Appointment struct {
	Id      int64 // unique identifier
	Time    int64 // unix time for easy sorting via int value comparison
	Patient *Patient
	Doctor  *Doctor
}

func getAppointmentsFromDB() ([]*Appointment, error) {

	rows, rowsErr := clinicDb.Query("SELECT * FROM appointment ORDER BY id ASC")

	if rowsErr != nil {
		return Appointments, ErrDBConn
	}

	for rows.Next() {

		var (
			id, time              int64
			doctor_id, patient_id string
		)

		rowScanErr := rows.Scan(&id, &time, &doctor_id, &patient_id)

		if rowScanErr != nil {
			return Appointments, ErrDBConn
		}

		doctorID, _ := strconv.ParseInt(doctor_id, 10, 64)
		doc, docErr := DoctorsBST.GetDoctorByIDBST(doctorID)
		pat, patErr := GetPatientByID(patient_id)

		if docErr == nil && patErr == nil {
			appt := &Appointment{
				id, time, pat, doc,
			}

			var appWg sync.WaitGroup

			appWg.Add(3)
			go addAppointment(appt, &appWg)
			go appt.Doctor.addAppointment(appt, &appWg)
			go appt.Patient.addAppointment(appt, &appWg)
			appWg.Wait()
		}
	}

	return Appointments, nil
}

// Create Appointment, insert to database, add Appointment to global slice AppointmentsSortedByTimeslot and Appointments, sort global slice AppointmentsSortedByTimeslot, patient's Appointments slice and doctor's Appointments slice by appointment time.
func MakeAppointment(t int64, pat *Patient, doc *Doctor, wgrp *sync.WaitGroup) (*Appointment, error) {

	if wgrp != nil {
		defer wgrp.Done()
	}

	mutex.Lock()
	defer mutex.Unlock()

	i := BinarySearchApptTime(t)

	if i >= 0 {
		return nil, ErrInvalidTimeslot
	}
	// Db
	stmt, prepErr := clinicDb.Prepare("INSERT into appointment (time, doctor_id, patient_id) values(?,?,?)")
	if prepErr != nil {
		log.Fatal(ErrDBConn.Error(), prepErr)
		return nil, ErrCreateAppointment
	}
	res, execErr := stmt.Exec(t, doc.Id, pat.Id)
	if execErr != nil {
		log.Fatal(ErrDBConn.Error(), execErr)
		return nil, ErrCreateAppointment
	}
	insertedId, insertedErr := res.LastInsertId()
	if insertedErr != nil {
		log.Fatal(ErrDBConn.Error(), insertedErr)
		return nil, ErrCreateAppointment
	}

	app := &Appointment{insertedId, t, pat, doc}

	var appWg sync.WaitGroup

	appWg.Add(3)
	go addAppointment(app, &appWg)
	go app.Doctor.addAppointment(app, &appWg)
	go app.Patient.addAppointment(app, &appWg)
	appWg.Wait()

	fmt.Println("Created Appt:", app.Id)

	return app, nil
}

func addAppointment(appt *Appointment, appWg *sync.WaitGroup) {
	defer appWg.Done()
	Appointments = append(Appointments, appt)
	sortAppointments()
}

// EditAppointment updates appointment item, updates corresponding database entry, sort AppointmentsSortedByTimeslot slice by time, Appointments slice by Id, patient's Appointments slice and doctor's Appointments slice by time.
func (appt *Appointment) EditAppointment(t int64, pat *Patient, doc *Doctor) error {

	mutex.Lock()
	{
		// Db
		_, execErr := clinicDb.Exec("UPDATE `appointment` SET time = ?, patient_id = ?, doctor_id = ? WHERE id = ?", t, pat.Id, doc.Id, appt.Id)
		if execErr != nil {
			log.Fatal(ErrDBConn.Error(), execErr)
		}

		// Update
		appt.Patient = pat
		appt.Doctor = doc
		appt.Time = t

		// Re-sort
		sortAppointments()
		pat.sortAppointments()
		doc.sortAppointments()
	}
	mutex.Unlock()

	return nil
}

// CancelAppointment removes appointment from AppointmentsSortedByTimeslot and Appointments, patient's Appointments slice, doctor's Appointments slice,
// delete corresponding database entry,
// sort AppointmentsSortedByTimeslot slice by time, Appointments slice by Id, patient's Appointments slice and doctor's Appointments slice by time.
func (appt *Appointment) CancelAppointment() {

	mutex.Lock()
	defer mutex.Unlock()

	apptIDIndex := BinarySearchApptID(appt.Id)

	if apptIDIndex >= 0 {

		// Db
		_, execErr := clinicDb.Exec("DELETE FROM `appointment` WHERE id = ?", appt.Id)
		if execErr != nil {
			log.Fatal(ErrDBConn.Error(), execErr)
		}

		var wg sync.WaitGroup

		// Remove from Patient & Doctor
		wg.Add(2)
		go Appointments[apptIDIndex].Patient.cancelAppointment(appt.Id, &wg)
		go Appointments[apptIDIndex].Doctor.cancelAppointment(appt.Id, &wg)
		wg.Wait()

		if apptIDIndex == 0 {
			Appointments = Appointments[1:]
		} else if apptIDIndex == len(Appointments)-1 {
			Appointments = Appointments[:apptIDIndex]
		} else {
			Appointments = append(Appointments[:apptIDIndex], Appointments[apptIDIndex+1:]...)
		}

		// Re-sort
		sortAppointments()
	}
}

// IsApptTimeValid checks if time of appointment is in the past - e.g. process started at 3:55 PM, user chose 4 PM timeslot but submitted form at 4:05 PM.
func IsApptTimeValid(t int64) (bool, error) {

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

	return true, nil
}

// IsThereTimeslot checks if there's timeslot available for the day by checking both the patient's and doctor's appointments for the day.
func IsThereTimeslot(dt int64, pat *Patient, doc *Doctor) (bool, error) {

	patientTimeslotsAvailable := GetAvailableTimeslot(dt, pat.GetAppointmentsByDate(dt))

	if len(patientTimeslotsAvailable) <= 0 {
		return false, ErrPatientNoMoreTimeslot
	}

	doctorTimeslotsAvailable := GetAvailableTimeslot(dt, doc.GetAppointmentsByDate(dt))

	if len(doctorTimeslotsAvailable) <= 0 {
		return false, ErrDoctorNoMoreTimeslot
	}

	timeslotsAvailable := GetAvailableTimeslot(dt, append(doc.GetAppointmentsByDate(dt), pat.GetAppointmentsByDate(dt)...))

	if len(timeslotsAvailable) <= 0 {
		return false, ErrNoMoreTimeslot
	}

	return true, nil
}

func sortAppointments() {
	updateIdSortedAppts()
	updateTimeslotSortedAppts()
}

func updateTimeslotSortedAppts() {
	tempAppts := make([]*Appointment, len(Appointments))
	copy(tempAppts, Appointments)
	mergeSortByTime(tempAppts, 0, len(tempAppts)-1)
	AppointmentsSortedByTimeslot = tempAppts
}

func updateIdSortedAppts() {
	mergeSortByAppointmentId(Appointments, 0, len(Appointments)-1)
}

// GetAvailableTimeslot returns a slice of all the possible open timeslots for a given day by getting the delta between all timeslots for the day and a slice of appointments on the day.
func GetAvailableTimeslot(dt int64, apptsToExclude []*Appointment) []int64 {

	allTimeSlots := timeSlotsGenerator(dt)

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

// timeSlotsGenerator returns slice of available time slots in 30 mins intervals from provided datetime.
func timeSlotsGenerator(dt int64) []int64 {

	selectedDate := time.Unix(dt, 0)

	currentTimeHour := time.Now().Hour()
	currentTimeMinute := time.Now().Minute()

	timeSlots := []int64{}
	startHour := currentTimeHour
	startMinute := 0
	currTime := time.Now()

	// Same day
	if selectedDate.Year() == currTime.Year() && selectedDate.Day() == currTime.Day() && selectedDate.Month() == currTime.Month() {
		if currentTimeHour >= startOperationHour {
			if currentTimeMinute >= 0 {
				startMinute = appointmentIntervals
			}
		} else {
			startHour = startOperationHour
		}
	} else {
		startHour = startOperationHour
		currTime = time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), startHour, startMinute, 0, 0, time.Local)
	}

	for startHour >= startOperationHour && startHour <= endOperationHour {

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

func mergeSortByAppointmentId(arr []*Appointment, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2                   // index of midpoint
		mergeSortByAppointmentId(arr, first, mid)   // sort left half
		mergeSortByAppointmentId(arr, mid+1, last)  // sort right half
		mergeByAppointmentId(arr, first, mid, last) // merge the two halves
	}
}

func mergeByAppointmentId(arr []*Appointment, first int, mid int, last int) {

	tempArr := make([]*Appointment, len(arr))

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

func mergeSortByTime(arr []*Appointment, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2          // index of midpoint
		mergeSortByTime(arr, first, mid)   // sort left half
		mergeSortByTime(arr, mid+1, last)  // sort right half
		mergeByTime(arr, first, mid, last) // merge the two halves
	}
}

func mergeByTime(arr []*Appointment, first int, mid int, last int) {

	tempArr := make([]*Appointment, len(arr))

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

// BinarySearchApptID performs binary search for appointment id in Appointments.
func BinarySearchApptID(apptID int64) int {
	return binarySearchAppt(Appointments, 0, len(Appointments)-1, apptID)
}

func binarySearchAppt(arr []*Appointment, first int, last int, apptID int64) int {
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

// BinarySearchApptTime performs binary search for appointment time in Appointments.
func BinarySearchApptTime(time int64) int {
	return binarySearchApptbyTime(AppointmentsSortedByTimeslot, 0, len(AppointmentsSortedByTimeslot)-1, time)
}

func binarySearchApptbyTime(arr []*Appointment, first int, last int, time int64) int {
	if first > last { // item not found
		return -1
	} else {
		mid := (first + last) / 2

		if arr[mid].Time == time { // item found
			return mid
		} else {
			if time < arr[mid].Time { // item in first half
				return binarySearchApptbyTime(arr, first, mid-1, time) // search in first half
			} else { // item in second half
				return binarySearchApptbyTime(arr, mid+1, last, time) // search in second half
			}
		}
	}
}
