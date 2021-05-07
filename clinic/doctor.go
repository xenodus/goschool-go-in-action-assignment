package clinic

import (
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Doctors hold all the doctors sorted by incremental id.
var Doctors = []*Doctor{}

// DoctorsBST is a balanced binary search tree of doctors.
var DoctorsBST *BST

type Doctor struct {
	Id           int64
	First_name   string
	Last_name    string
	Appointments []*Appointment
}

func getDoctorsFromDB() ([]*Doctor, error) {

	rows, rowsErr := clinicDb.Query("SELECT * FROM doctor ORDER BY id ASC")

	if rowsErr != nil {
		return Doctors, ErrDBConn
	}

	for rows.Next() {

		var (
			id                    int64
			first_name, last_name string
		)

		rowScanErr := rows.Scan(&id, &first_name, &last_name)

		if rowScanErr != nil {
			return Doctors, ErrDBConn
		}

		doc := &Doctor{
			id, first_name, last_name, nil,
		}

		Doctors = append(Doctors, doc)
	}

	if len(Doctors) > 0 {
		DoctorsBST = makeBST()
	}

	return Doctors, nil
}

func addDoctor(first_name string, last_name string) (*Doctor, error) {
	defer Wg.Done()

	mutex.Lock()
	defer mutex.Unlock()

	stmt, prepErr := clinicDb.Prepare("INSERT into doctor (first_name, last_name) values(?,?)")
	if prepErr != nil {
		log.Fatal(ErrDBConn.Error(), prepErr)
		return nil, ErrCreateDoctor
	}
	res, execErr := stmt.Exec(first_name, last_name)
	if execErr != nil {
		log.Fatal(ErrDBConn.Error(), execErr)
		return nil, ErrCreateDoctor
	}
	insertedId, insertedErr := res.LastInsertId()
	if insertedErr != nil {
		log.Fatal(ErrDBConn.Error(), insertedErr)
		return nil, ErrCreateDoctor
	}

	doc := &Doctor{insertedId, first_name, last_name, nil}
	Doctors = append(Doctors, doc)
	sortDoctorsById()
	DoctorsBST = makeBST()

	return doc, nil
}

func (d *Doctor) IsFreeAt(t int64) bool {
	for _, v := range d.Appointments {
		if v.Time == t {
			return false
		}
	}
	return true
}

// GetAppointmentsByDate returns a slice of Appointments (pointers) on the given date (unix time).
// Todo: Can improve by making it binary search instead of sequential since Appointments is sorted by time.
func (d *Doctor) GetAppointmentsByDate(dt int64) []*Appointment {

	mutex.Lock()
	defer mutex.Unlock()

	requestedDateTime := time.Unix(dt, 0)
	appts := []*Appointment{}

	for _, v := range d.Appointments {
		apptDateTime := time.Unix(v.Time, 0)
		if apptDateTime.Year() == requestedDateTime.Year() && apptDateTime.Month() == requestedDateTime.Month() && apptDateTime.Day() == requestedDateTime.Day() {
			appts = append(appts, v)
		}
	}

	return appts
}

func (d *Doctor) sortAppointments() {
	mergeSortByTime(d.Appointments, 0, len(d.Appointments)-1) // sorted by time
}

func (d *Doctor) addAppointment(appt *Appointment) {
	d.Appointments = append(d.Appointments, appt)
	d.sortAppointments()
}

func (d *Doctor) cancelAppointment(apptID int64, wg *sync.WaitGroup) error {
	defer wg.Done()

	apptIDIndex, err := d.searchApptID(apptID)

	if apptIDIndex >= 0 {

		if apptIDIndex == 0 {
			d.Appointments = d.Appointments[1:]
		} else if apptIDIndex == len(d.Appointments)-1 {
			d.Appointments = d.Appointments[:apptIDIndex]
		} else {
			d.Appointments = append(d.Appointments[:apptIDIndex], d.Appointments[apptIDIndex+1:]...)
		}
	}

	return err
}

type BinaryNode struct {
	doctor *Doctor     // to store the data
	left   *BinaryNode // pointer to point to left node
	right  *BinaryNode // pointer to point to right node
}

type BST struct {
	root *BinaryNode
}

func arrayToBinaryTree(a []*Doctor, start int, end int) *BinaryNode {
	if start > end {
		var returnNode *BinaryNode
		return returnNode
	}

	middle := (start + end) / 2
	node := BinaryNode{doctor: a[middle]}
	node.left = arrayToBinaryTree(a, start, middle-1)
	node.right = arrayToBinaryTree(a, middle+1, end)

	return &node
}

func makeBST() *BST {
	bn := arrayToBinaryTree(Doctors, 0, len(Doctors)-1)
	DoctorBST := BST{bn}
	return &DoctorBST
}

/*
func (bst *BST) printDoctorsInOrder() {
	bst.inOrderTraverse(bst.root)
}

func (bst *BST) inOrderTraverse(t *BinaryNode) {
	if t != nil {
		bst.inOrderTraverse(t.left)
		fmt.Println(strconv.FormatInt(t.doctor.Id, 10)+".", t.doctor.First_name, t.doctor.Last_name)
		bst.inOrderTraverse(t.right)
	}
}
*/

func (bst *BST) searchNode(t *BinaryNode, docID int64) *BinaryNode {
	if t == nil {
		return nil
	} else {
		if t.doctor.Id == docID {
			return t
		} else {
			if docID < t.doctor.Id {
				return bst.searchNode(t.left, docID)
			} else {
				return bst.searchNode(t.right, docID)
			}
		}
	}
}

// GetDoctorByIDBST gets a Doctor from the global DoctorsBST by Id. Returns pointer to Doctor if found.
func (bst *BST) GetDoctorByIDBST(docID int64) (*Doctor, error) {

	docBN := bst.searchNode(bst.root, docID)

	if docBN != nil {
		return docBN.doctor, nil
	}

	return nil, ErrDoctorIDNotFound
}

// Sequential search
func (d *Doctor) searchApptID(apptID int64) (int, error) {
	for k, v := range d.Appointments {
		if v.Id == apptID {
			return k, nil
		}
	}
	return -1, ErrAppointmentIDNotFound
}

func sortDoctorsById() {
	mergeSortByDoctorId(Doctors, 0, len(Doctors)-1)
}

func mergeSortByDoctorId(arr []*Doctor, first int, last int) {
	if first < last { // more than 1 items
		mid := (first + last) / 2              // index of midpoint
		mergeSortByDoctorId(arr, first, mid)   // sort left half
		mergeSortByDoctorId(arr, mid+1, last)  // sort right half
		mergeByDoctorId(arr, first, mid, last) // merge the two halves
	}
}

func mergeByDoctorId(arr []*Doctor, first int, mid int, last int) {

	tempArr := make([]*Doctor, len(arr))

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
