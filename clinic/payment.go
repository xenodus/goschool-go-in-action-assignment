package clinic

import (
	"strconv"
	"strings"
	"sync/atomic"
)

// Globals
var PaymentQ = PaymentQueue{}
var MissedPaymentQ = PaymentQueue{}
var paymentCounter int64 = 200

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

func CreatePayment(appt *Appointment, amt float64) error {

	atomic.AddInt64(&paymentCounter, 1)
	pmy := Payment{paymentCounter, appt, amt}
	PaymentQ.Enqueue(&pmy)
	appt.CancelAppointment()

	return nil
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
