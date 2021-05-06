package web

// Routes
const (
	pageIndex    = "/"
	pageLogin    = "/login"
	pageLogout   = "/logout"
	pageRegister = "/register"
	pageProfile  = "/profile"

	pageError = "/error"

	pageMyAppointments  = "/appointments"
	pageNewAppointment  = "/appointment/new"
	pageEditAppointment = "/appointment/edit"

	pageDoctors = "/doctors"

	pageAdminDebug           = "/admin/debug"
	pageAdminAllAppointments = "/admin/appointments"
	pageAdminEditAppointment = "/admin/appointment/edit"
	pageAdminSessions        = "/admin/sessions"
	pageAdminUsers           = "/admin/users"
	pageAdminPaymentEnqueue  = "/admin/payment/enqueue"

	pageAdminPaymentDequeue               = "/admin/payment/dequeue"
	pageAdminPaymentDequeueToMissedQueue  = "/admin/payment/dequeueToMissedQueue"
	pageAdminPaymentDequeueToPaymentQueue = "/admin/payment/dequeueToPaymentQueue"

	pagePaymentQueue = "/payment/queue"
	pagePSI          = "/psi"
)
