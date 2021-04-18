package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var paymentCounter int64 = 200

type payment struct {
	Id          int64
	Appointment *appointment
	Amount      float64
}

type paymentNode struct {
	Payment *payment
	Next    *paymentNode
}

type paymentQueue struct {
	Front *paymentNode
	Back  *paymentNode
	Size  int
}

func (p *paymentQueue) enqueue(pmy *payment) error {
	newNode := &paymentNode{
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

func (p *paymentQueue) dequeue() (*payment, error) {
	var pmy *payment

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

func (p *paymentQueue) PrintAllQueueIDs(skipFirst bool) string {

	queueIds := p.getAllQueueID()

	fmt.Println(queueIds)

	if len(queueIds) > 0 {
		if skipFirst {
			return strings.Join(queueIds[1:], ", ")
		} else {
			return strings.Join(queueIds, ", ")
		}
	}

	return ""
}

func (p *paymentQueue) getAllQueueID() []string {
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
func (p *paymentQueue) dequeueToMissedPaymentQueue() (*payment, error) {
	pmy, err := p.dequeue()

	if err == nil {
		missedPaymentQ.enqueue(pmy)
		return pmy, nil
	}

	return nil, err
}

// Move from missed payment queue back to main payment queue
func (p *paymentQueue) dequeueToPaymentQueue() (*payment, error) {
	pmy, err := p.dequeue()

	if err == nil {
		paymentQ.enqueue(pmy)
		return pmy, nil
	}

	return nil, err
}

func paymentQueuePage(res http.ResponseWriter, req *http.Request) {

	var theUser *patient

	if isLoggedIn(req) {
		theUser = getLoggedInPatient(res, req)
	}

	// Anonymous payload
	payload := struct {
		Queue       *paymentQueue
		MissedQueue *paymentQueue
		User        *patient
	}{
		&paymentQ,
		&missedPaymentQ,
		theUser,
	}

	tpl.ExecuteTemplate(res, "paymentQueue.gohtml", payload)
}
