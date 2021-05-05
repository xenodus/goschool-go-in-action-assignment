package clinic

import (
	"log"
	"strconv"
	"strings"
)

// Globals
var PaymentQ = &PaymentQueue{}
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

func CreatePayment(appt *Appointment, amt float64) (*PaymentQueue, error) {

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
	appt.CancelAppointment()

	return PaymentQ, nil
}

func (pmy *Payment) ClearPayment() {
	// Db
	_, execErr := clinicDb.Exec("DELETE FROM `payment` WHERE id = ?", pmy.Id)
	if execErr != nil {
		log.Fatal(ErrDBConn.Error(), execErr)
	}
}

func (p *PaymentQueue) Enqueue(pmy *Payment) error {

	mutex.Lock()
	{
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
	}
	mutex.Unlock()

	return nil
}

func (p *PaymentQueue) Dequeue() (*Payment, error) {

	var pmy *Payment

	mutex.Lock()
	{
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
	}
	mutex.Unlock()

	return pmy, nil
}

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

// Move over to missed queues if say nobody turns up
func (p *PaymentQueue) DequeueToMissedPaymentQueue() (*Payment, error) {
	pmy, err := p.Dequeue()

	if err == nil {
		MissedPaymentQ.Enqueue(pmy)
		return pmy, nil
	}

	return nil, err
}

// Move from missed payment queue back to main payment queue
func (p *PaymentQueue) DequeueToPaymentQueue() (*Payment, error) {
	pmy, err := p.Dequeue()

	if err == nil {
		PaymentQ.Enqueue(pmy)
		return pmy, nil
	}

	return nil, err
}
