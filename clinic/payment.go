package clinic

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

// PaymentQ holds the payments that are pending in a FIFO queue.
var PaymentQ = &PaymentQueue{}

// MissedPaymentQ holds the outstanding payments that have been moved over from PaymentQ.
var MissedPaymentQ = &PaymentQueue{}

type Payment struct {
	Id          int64
	Appointment *Appointment
	Amount      float64
}

type PaymentNode struct {
	Payment *Payment
	Next    *PaymentNode
}

type PaymentQueue struct {
	Front *PaymentNode
	Back  *PaymentNode
	Size  int
}

func getPaymentsFromDB() (*PaymentQueue, error) {

	rows, rowsErr := clinicDb.Query("SELECT * FROM payment")

	if rowsErr != nil {
		return PaymentQ, ErrDBConn
	}

	for rows.Next() {

		var (
			id, appointment_id int64
			amount             float64
		)

		rowScanErr := rows.Scan(&id, &amount, &appointment_id)

		if rowScanErr != nil {
			return PaymentQ, ErrDBConn
		}

		appt := Appointment{appointment_id, 0, nil, nil}
		pmy := Payment{id, &appt, amount}
		PaymentQ.Enqueue(&pmy)
	}

	return PaymentQ, nil
}

// Create payment, add to database, add to payment queue and remove the appointment.
func CreatePayment(appt *Appointment, amt float64, wg *sync.WaitGroup) (*PaymentQueue, error) {

	if wg != nil {
		defer wg.Done()
	}

	mutex.Lock()
	{

		// Db
		stmt, prepErr := clinicDb.Prepare("INSERT into payment (amount, appointment_id) values(?,?)")
		if prepErr != nil {
			log.Fatal(ErrDBConn.Error(), prepErr)
			return PaymentQ, ErrCreateAppointment
		}
		res, execErr := stmt.Exec(amt, appt.Id)
		if execErr != nil {
			log.Fatal(ErrDBConn.Error(), execErr)
			return PaymentQ, ErrCreateAppointment
		}
		insertedId, insertedErr := res.LastInsertId()
		if insertedErr != nil {
			log.Fatal(ErrDBConn.Error(), insertedErr)
			return PaymentQ, ErrCreateAppointment
		}

		pmy := Payment{insertedId, appt, amt}
		PaymentQ.Enqueue(&pmy)

		fmt.Println("Payment Q enqueued:", pmy.Appointment.Id)
	}
	mutex.Unlock()

	return PaymentQ, nil
}

// Delete payment entry from database.
func (pmy *Payment) ClearPayment() {
	// Db
	_, execErr := clinicDb.Exec("DELETE FROM `payment` WHERE id = ?", pmy.Id)
	if execErr != nil {
		log.Fatal(ErrDBConn.Error(), execErr)
	}
}

// Add payment to a queue.
func (p *PaymentQueue) Enqueue(pmy *Payment) error {

	newNode := &PaymentNode{
		Payment: pmy,
		Next:    nil,
	}

	if p.Front == nil {
		p.Front = newNode
	} else {
		p.Back.Next = newNode
	}

	p.Back = newNode
	p.Size++

	return nil
}

// Remove payment from a queue.
func (p *PaymentQueue) Dequeue() (*Payment, error) {

	var pmy *Payment

	if p.Front == nil {
		return nil, ErrEmptyPaymentQueue
	}

	pmy = p.Front.Payment

	if p.Size == 1 {
		p.Front = nil
		p.Back = nil
	} else {
		p.Front = p.Front.Next
	}

	p.Size--

	return pmy, nil
}

// Returns a CSV concatenated string of appointment ids from payments inside a payment queue.
func (p *PaymentQueue) PrintAllQueueIDs(skipFirst bool) string {

	queueIds := p.getAllQueueID()

	if len(queueIds) > 0 {
		if skipFirst {
			return strings.Join(queueIds[1:], ", ")
		} else {
			return strings.Join(queueIds, ", ")
		}
	}

	return ""
}

func (p *PaymentQueue) getAllQueueID() []string {
	queueIDs := []string{}
	currentNode := p.Front

	if currentNode == nil {
		return queueIDs
	}

	queueIDs = append(queueIDs, strconv.FormatInt(currentNode.Payment.Appointment.Id, 10))

	for currentNode.Next != nil {
		currentNode = currentNode.Next
		queueIDs = append(queueIDs, strconv.FormatInt(currentNode.Payment.Appointment.Id, 10))
	}

	return queueIDs
}

// Remove a payment from a queue and move it to MissedPaymentQ.
func (p *PaymentQueue) DequeueToMissedPaymentQueue() (*Payment, error) {

	mutex.Lock()
	defer mutex.Unlock()

	pmy, err := p.Dequeue()

	if err == nil {
		MissedPaymentQ.Enqueue(pmy)
		return pmy, nil
	}

	return nil, err
}

// Remove a payment from a queue and move it to PaymentQ.
func (p *PaymentQueue) DequeueToPaymentQueue() (*Payment, error) {

	mutex.Lock()
	defer mutex.Unlock()

	pmy, err := p.Dequeue()

	if err == nil {
		PaymentQ.Enqueue(pmy)
		return pmy, nil
	}

	return nil, err
}
