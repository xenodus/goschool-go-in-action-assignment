package clinic

import (
	"sync"
	"sync/atomic"
	"time"
)

// Globals
var Appointments = []*Appointment{}
var AppointmentsSortedByTimeslot = []*Appointment{}
var appointment_start_id int64 = 1000

var wg sync.WaitGroup
var mutex sync.Mutex

type Appointment struct {
	Id      int64 // unique identifier
	Time    int64 // unix time for easy sorting via int value comparison
	Patient *Patient
	Doctor  *Doctor
}

// Make and sort by appointment time
func MakeAppointment(t int64, pat *Patient, doc *Doctor) (*Appointment, error) {

	mutex.Lock()
	defer mutex.Unlock()

	app := Appointment{}
	_, err := IsThereTimeslot(pat, doc)

	if err == nil {

		atomic.AddInt64(&appointment_start_id, 1)
		app := Appointment{appointment_start_id, t, pat, doc}

		wg.Add(3)
		go addAppointment(&app)
		go app.Doctor.addAppointment(&app)
		go app.Patient.addAppointment(&app)
		wg.Wait()

		return &app, nil
	}

	return &app, err
}

func addAppointment(appt *Appointment) {
	defer wg.Done()
	Appointments = append(Appointments, appt)
	updateTimeslotSortedAppts()
}

func (appt *Appointment) EditAppointment(t int64, pat *Patient, doc *Doctor) error {

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

func (appt *Appointment) CancelAppointment() {
	mutex.Lock()
	{
		apptIDIndex := BinarySearchApptID(appt.Id)

		if apptIDIndex >= 0 {
			// Remove from Patient & Doctor
			wg.Add(2)
			go Appointments[apptIDIndex].Patient.cancelAppointment(appt.Id)
			go Appointments[apptIDIndex].Doctor.cancelAppointment(appt.Id)
			wg.Wait()

			if apptIDIndex == 0 {
				Appointments = Appointments[1:]
			} else if apptIDIndex == len(Appointments)-1 {
				Appointments = Appointments[:apptIDIndex]
			} else {
				Appointments = append(Appointments[:apptIDIndex], Appointments[apptIDIndex+1:]...)
			}

			// Re-sort appointmentsSortedByTimeslot by time
			updateTimeslotSortedAppts()
		}
	}
	mutex.Unlock()
}

// Check if time of appointment is in the past - e.g. process started at 3:55 PM, user chose 4 PM timeslot but submitted at 4:05 PM
func IsApptTimeValid(t int64) (bool, error) {

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

func IsThereTimeslot(pat *Patient, doc *Doctor) (bool, error) {

	patientTimeslotsAvailable := GetAvailableTimeslot(pat.Appointments)

	if len(patientTimeslotsAvailable) <= 0 {
		return false, ErrPatientNoMoreTimeslot
	}

	doctorTimeslotsAvailable := GetAvailableTimeslot(doc.Appointments)

	if len(doctorTimeslotsAvailable) <= 0 {
		return false, ErrDoctorNoMoreTimeslot
	}

	timeslotsAvailable := GetAvailableTimeslot(append(doc.Appointments, pat.Appointments...))

	if len(timeslotsAvailable) <= 0 {
		return false, ErrNoMoreTimeslot
	}

	return true, nil
}

func updateTimeslotSortedAppts() {
	tempAppts := make([]*Appointment, len(Appointments))
	copy(tempAppts, Appointments)
	mergeSortByTime(tempAppts, 0, len(tempAppts)-1)
	AppointmentsSortedByTimeslot = tempAppts
}

func GetAvailableTimeslot(apptsToExclude []*Appointment) []int64 {

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

// Binary search for appointment id in sorted slice
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

// Sequential search
func searchApptID(arr []*Appointment, apptID int64) (int, error) {
	for k, v := range arr {
		if v.Id == apptID {
			return k, nil
		}
	}
	return -1, ErrAppointmentIDNotFound
}
