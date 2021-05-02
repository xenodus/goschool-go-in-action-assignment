package psi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type PSI struct {
	Value       string
	Description string
}

func GetPSI() (*PSI, error) {

	var jsonResult map[string]interface{}

	api_url := "https://api.data.gov.sg/v1/environment/psi"

	result, httpErr := http.Get(api_url)

	if httpErr != nil {
		return nil, errors.New("PSI fetch failure: unable to retrieve json endpoint, " + api_url)
	}

	JSONData, _ := ioutil.ReadAll(result.Body)
	marshalErr := json.Unmarshal(JSONData, &jsonResult)
	tmp := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"].(map[string]interface{})["national"]

	if marshalErr != nil || tmp == nil {
		return nil, errors.New("PSI fetch failure: error decoding json")
	}

	psi := fmt.Sprintf("%v", tmp)
	psiInt, err := strconv.Atoi(psi)

	if err != nil {
		return nil, errors.New("PSI fetch failure: error parsing psi to int")
	}

	// All good here
	newPSI := PSI{
		Value:       psi,
		Description: "",
	}

	// https://www.haze.gov.sg/ for groupings
	if psiInt >= 0 && psiInt <= 55 {
		newPSI.Description = "Normal"
	} else if psiInt >= 56 && psiInt <= 150 {
		newPSI.Description = "Elevated"
	} else if psiInt >= 151 && psiInt <= 250 {
		newPSI.Description = "High"
	} else {
		newPSI.Description = "Very High"
	}

	return &newPSI, nil
}
