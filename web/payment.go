package web

import (
	"assignment4/clinic"
	"net/http"
)

func paymentQueuePage(res http.ResponseWriter, req *http.Request) {

	thePatient, _ := clinic.IsLoggedIn(req)

	// Anonymous payload
	payload := struct {
		PageTitle   string
		Queue       *clinic.PaymentQueue
		MissedQueue *clinic.PaymentQueue
		User        *clinic.Patient
	}{
		"Payment Queue",
		clinic.PaymentQ,
		clinic.MissedPaymentQ,
		thePatient,
	}

	tpl.ExecuteTemplate(res, "paymentQueue.gohtml", payload)
}
