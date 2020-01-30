package ssllabs

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// APIInfo /info endpoint result
type APIInfo struct {
	EngineVersion      string `json:"engineVersion"`
	CriteriaVersion    string `json:"criteriaVersion"`
	MaxAssessments     int    `json:"maxAssessments"`
	CurrentAssessments int    `json:"currentAssessments"`
}

// Info calls /info endpoint and returns and Info
func Info() (info APIInfo, err error) {
	response, err := http.Get(API + "/info")
	if err != nil {
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return
	}

	return
}
