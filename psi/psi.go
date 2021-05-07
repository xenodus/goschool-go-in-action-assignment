// Package psi defines the PSI type and provide implementation for fetching the 24H average Pollutant Standards Index (PSI) value.
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

// GetPSI returns a PSI item containing the 24H national average pollutant standards index value and description.
func GetPSI() (*PSI, error) {

	var jsonResult map[string]interface{}

	api_url := "https://api.data.gov.sg/v1/environment/psi"

	result, httpErr := http.Get(api_url)

	if httpErr != nil {
		return nil, errors.New("PSI fetch failure: unable to retrieve json endpoint," + api_url)
	}

	JSONData, _ := ioutil.ReadAll(result.Body)
	marshalErr := json.Unmarshal(JSONData, &jsonResult)

	if marshalErr != nil {
		return nil, errors.New("PSI fetch failure: error decoding json")
	}

	tmp, err := parseJSON(jsonResult)

	if err != nil {
		return nil, err
	}

	psi := fmt.Sprintf("%v", tmp)
	psiInt, err := strconv.Atoi(psi)

	if err != nil {
		return nil, errors.New("PSI fetch failure: error parsing psi to int")
	}

	newPSI := &PSI{
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

	return newPSI, nil
}

// Is there a better way to check / unmarshal json with many similarly named keys aside from traversing and checking each level?
func parseJSON(jsonResult map[string]interface{}) (interface{}, error) {

	if _, ok := jsonResult["items"].([]interface{}); !ok {
		return nil, errors.New("PSI fetch failure: error decoding json. Can't find key, items")
	} else if len(jsonResult["items"].([]interface{})) == 0 {
		return nil, errors.New("PSI fetch failure: error decoding json. Empty items")
	} else if _, ok := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"]; !ok {
		return nil, errors.New("PSI fetch failure: error decoding json. Can't find key, readings")
	} else if _, ok := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{}); !ok {
		return nil, errors.New("PSI fetch failure: error decoding json. Empty readings")
	} else if _, ok := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"]; !ok {
		return nil, errors.New("PSI fetch failure: error decoding json. Can't find key, psi_twenty_four_hourly")
	} else if _, ok := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"].(map[string]interface{}); !ok {
		return nil, errors.New("PSI fetch failure: error decoding json. Empty psi_twenty_four_hourly")
	} else if _, ok := jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"].(map[string]interface{})["national"]; !ok {
		return nil, errors.New("PSI fetch failure: error decoding json. Can't find key, national")
	}

	return jsonResult["items"].([]interface{})[0].(map[string]interface{})["readings"].(map[string]interface{})["psi_twenty_four_hourly"].(map[string]interface{})["national"], nil
}
