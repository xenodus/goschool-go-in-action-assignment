package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func psiPage(res http.ResponseWriter, req *http.Request) {

	thePatient, _ := isLoggedIn(req)

	// Anonymous payload
	payload := struct {
		User           *patient
		PageTitle      string
		ErrorMsg       string
		Psi            string
		PsiDescription string
	}{
		thePatient, "PSI", "", "", "",
	}

	var jsonResult map[string]interface{}

	api_url := "https://api.data.gov.sg/v1/environment/psi"

	result, httpErr := http.Get(api_url)

	if httpErr != nil {
		payload.ErrorMsg = "Error encountered fetching PSI"
	} else {
		JSONData, _ := ioutil.ReadAll(result.Body)
		marshalErr := json.Unmarshal(JSONData, &jsonResult)
		tmp := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"].(map[string]interface{})["national"]

		if marshalErr != nil || tmp == nil {
			payload.ErrorMsg = "Error encountered fetching PSI"
		} else {
			payload.Psi = fmt.Sprintf("%v", tmp)
			psiInt, err := strconv.Atoi(payload.Psi)

			if err == nil {
				// https://www.haze.gov.sg/
				if psiInt >= 0 && psiInt <= 55 {
					payload.PsiDescription = "Normal"
				} else if psiInt >= 56 && psiInt <= 150 {
					payload.PsiDescription = "Elevated"
				} else if psiInt >= 151 && psiInt <= 250 {
					payload.PsiDescription = "High"
				} else {
					payload.PsiDescription = "Very High"
				}
			}
		}
	}

	tpl.ExecuteTemplate(res, "psi.gohtml", payload)
}
