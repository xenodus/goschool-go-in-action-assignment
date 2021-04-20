package main

import (
	"net/http"
	"strconv"
)

var doctor_start_id int64 = 100

type doctor struct {
	Id           int64
	First_name   string
	Last_name    string
	Appointments []*appointment
}

func addDoctor(first_name string, last_name string) {
	defer wg.Done()

	mutex.Lock()
	{
		doctor_start_id++
		doc := doctor{doctor_start_id, first_name, last_name, nil}
		doctors = append(doctors, &doc)
		doctorsBST = makeBST()
	}
	mutex.Unlock()
}

func (d *doctor) isFreeAt(t int64) bool {
	for _, v := range d.Appointments {
		if v.Time == t {
			return false
		}
	}

	return true
}

func (d *doctor) sortAppointments() {
	mergeSort(d.Appointments, 0, len(d.Appointments)-1)
}

func (d *doctor) addAppointment(appt *appointment) {
	d.Appointments = append(d.Appointments, appt)
	d.sortAppointments()
}

func (d *doctor) cancelAppointment(apptID int64) error {
	defer wg.Done()

	apptIDIndex, err := searchApptID(d.Appointments, apptID)

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

// BST
type BinaryNode struct {
	doctor *doctor     // to store the data
	left   *BinaryNode // pointer to point to left node
	right  *BinaryNode // pointer to point to right node
}

type BST struct {
	root *BinaryNode
}

func arrayToBinaryTree(a []*doctor, start int, end int) *BinaryNode {
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
	bn := arrayToBinaryTree(doctors, 0, len(doctors)-1)
	doctorBST := BST{bn}
	return &doctorBST
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

func (bst *BST) getDoctorByIDBST(docID int64) (*doctor, error) {

	docBN := bst.searchNode(bst.root, docID)

	if docBN != nil {
		return docBN.doctor, nil
	}

	return nil, ErrDoctorIDNotFound
}

// Web Pages

func viewDoctorsPage(res http.ResponseWriter, req *http.Request) {

	if !isLoggedIn(req) {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Get querystring values
	doctorID := req.FormValue("doctorID")
	thePatient := getLoggedInPatient(res, req)

	var chosenDoctor *doctor = nil
	var timeslotsAvailable []int64
	var errorMsg = ""

	if doctorID != "" {
		doctorID, _ := strconv.ParseInt(doctorID, 10, 64)
		doc, err := doctorsBST.getDoctorByIDBST(doctorID)

		if err == nil {
			chosenDoctor = doc
			timeslotsAvailable = getAvailableTimeslot(chosenDoctor.Appointments)
		} else {
			errorMsg = err.Error()
		}
	}

	// Anonymous payload
	payload := struct {
		PageTitle          string
		ErrorMsg           string
		User               *patient
		ChosenDoctor       *doctor
		TimeslotsAvailable []int64
		Doctors            []*doctor
	}{
		"View Doctors",
		errorMsg,
		thePatient,
		chosenDoctor,
		timeslotsAvailable,
		doctors,
	}

	tpl.ExecuteTemplate(res, "doctors.gohtml", payload)
}
