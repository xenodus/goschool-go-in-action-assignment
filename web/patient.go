package web

import (
	"assignment4/clinic"
	"assignment4/session"
	"net/http"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/crypto/bcrypt"
)

// Globals
var wg sync.WaitGroup

// Web Pages
func areInputValid(username, firstname, lastname, password string, isRegister bool) error {

	if firstname == "" || lastname == "" || password == "" {
		return errMissingField
	}

	if len(password) < clinic.MinPasswordLength {
		return errPasswordLength
	}

	if !clinic.IsNRICValid(username) {
		return errInvalidNRIC
	}

	if isRegister {
		if _, err := clinic.GetPatientByID(username); err == nil {
			return errAlreadyRegistered
		}
	}

	return nil
}

func registerPage(res http.ResponseWriter, req *http.Request) {

	if _, isLoggedInCheck := clinic.IsLoggedIn(req); isLoggedInCheck {
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *clinic.Patient
		ErrorMsg  string
	}{
		"Register", nil, "",
	}

	// Process form submission
	if req.Method == http.MethodPost {

		// Policy to disallow and strip all tags - Similar to GO's unexported striptags
		p := bluemonday.StrictPolicy()

		username := p.Sanitize(strings.ToUpper(strings.TrimSpace(req.FormValue("nric"))))
		firstname := p.Sanitize(strings.TrimSpace(req.FormValue("firstname")))
		lastname := p.Sanitize(strings.TrimSpace(req.FormValue("lastname")))
		password := req.FormValue("password")

		inputErr := areInputValid(username, firstname, lastname, password, true)

		if inputErr != nil {
			payload.ErrorMsg = inputErr.Error()
			doLog(req, "WARNING", " Registration input validation failure")
		} else {
			// Create session + cookie
			session.CreateSession(res, req, username, serverHost)

			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				doLog(req, "ERROR", " Password bcrypt generation failure")
				http.Redirect(res, req, pageError+"?err=ErrInternalServerError", http.StatusSeeOther)
				return
			}

			wg.Add(1)
			clinic.CreatePatient(username, firstname, lastname, bPassword)
			wg.Wait()

			// Redirect to main index
			http.Redirect(res, req, pageIndex, http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(res, "register.gohtml", payload)
}

func loginPage(res http.ResponseWriter, req *http.Request) {

	if _, isLoggedInCheck := clinic.IsLoggedIn(req); isLoggedInCheck {
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle string
		User      *clinic.Patient
		ErrorMsg  string
	}{
		"Login", nil, "",
	}

	// Process form submission
	if req.Method == http.MethodPost {
		username := strings.ToUpper(strings.TrimSpace(req.FormValue("nric")))
		password := req.FormValue("password")
		// Check if user exist with username
		myUser, noPatientErr := clinic.GetPatientByID(username)

		if noPatientErr != nil {
			payload.ErrorMsg = errAuthFailure.Error()
			res.WriteHeader(http.StatusForbidden)
			doLog(req, "WARNING", " Login failure")
		}

		if payload.ErrorMsg == "" {
			// Matching of password entered
			err := bcrypt.CompareHashAndPassword(myUser.Password, []byte(password))
			if err != nil {
				payload.ErrorMsg = errAuthFailure.Error()
				res.WriteHeader(http.StatusForbidden)
				doLog(req, "WARNING", " Login failure")
			}
		}

		if payload.ErrorMsg == "" {
			// Create session + cookie
			session.CreateSession(res, req, username, serverHost)
			http.Redirect(res, req, pageIndex, http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(res, "login.gohtml", payload)
}

func logoutPage(res http.ResponseWriter, req *http.Request) {

	if _, isLoggedInCheck := clinic.IsLoggedIn(req); !isLoggedInCheck {
		http.Redirect(res, req, pageIndex, http.StatusSeeOther)
		return
	}

	// Delete session + cookie
	session.DeleteSession(res, req)

	http.Redirect(res, req, pageIndex, http.StatusSeeOther)
}

func profilePage(res http.ResponseWriter, req *http.Request) {

	thePatient, isLoggedInCheck := clinic.IsLoggedIn(req)

	if !isLoggedInCheck {
		http.Redirect(res, req, pageLogin, http.StatusSeeOther)
		return
	}

	// Anonymous payload
	payload := struct {
		PageTitle  string
		User       *clinic.Patient
		ErrorMsg   string
		SuccessMsg string
	}{
		"Profile", thePatient, "", "",
	}

	// Form submit values
	if req.Method == http.MethodPost {

		// Policy to disallow and strip all tags - Similar to GO's unexported striptags
		p := bluemonday.StrictPolicy()

		firstname := p.Sanitize(strings.TrimSpace(req.FormValue("firstname")))
		lastname := p.Sanitize(strings.TrimSpace(req.FormValue("lastname")))
		password := req.FormValue("password")

		inputErr := areInputValid(thePatient.Id, firstname, lastname, password, false)

		if inputErr != nil {
			payload.ErrorMsg = inputErr.Error()
			doLog(req, "WARNING", " Profile update input validation failure")
		} else {
			bPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			thePatient.EditPatient(thePatient.Id, firstname, lastname, bPassword)
			payload.SuccessMsg = "Profile updated!"
		}
	}

	tpl.ExecuteTemplate(res, "profile.gohtml", payload)
}
