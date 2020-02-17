// Copyright 2020 Anas Ait Said Oubrahim

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
