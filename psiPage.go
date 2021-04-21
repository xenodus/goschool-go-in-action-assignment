package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func psiPage(res http.ResponseWriter, req *http.Request) {

	var psi = ""
	var psiDescription = ""
	var errorMsg = ""
	var jsonResult map[string]interface{}

	api_url := "https://api.data.gov.sg/v1/environment/psi"

	result, httpErr := http.Get(api_url)

	if httpErr != nil {
		errorMsg = "Error encountered fetching PSI"
	} else {
		JSONData, _ := ioutil.ReadAll(result.Body)
		marshalErr := json.Unmarshal(JSONData, &jsonResult)
		tmp := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"].(map[string]interface{})["national"]

		if marshalErr != nil || tmp == nil {
			errorMsg = "Error encountered fetching PSI"
		} else {
			psi = fmt.Sprintf("%v", tmp)
			psiInt, err := strconv.Atoi(psi)

			if err == nil {
				// https://www.haze.gov.sg/
				if psiInt >= 0 && psiInt <= 55 {
					psiDescription = "Normal"
				} else if psiInt >= 56 && psiInt <= 150 {
					psiDescription = "Elevated"
				} else if psiInt >= 151 && psiInt <= 250 {
					psiDescription = "High"
				} else {
					psiDescription = "Very High"
				}
			}
		}
	}

	// Anonymous payload
	payload := struct {
		User           *patient
		PageTitle      string
		ErrorMsg       string
		Psi            string
		PsiDescription string
	}{
		nil,
		"PSI",
		errorMsg,
		psi,
		psiDescription,
	}

	tpl.ExecuteTemplate(res, "psi.gohtml", payload)
}
