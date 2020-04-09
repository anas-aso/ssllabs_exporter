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

package main

import (
	"net/http"
	"testing"
	"time"
)

func TestGetTimeout(t *testing.T) {
	var cases = []struct {
		name              string
		flagTimeout       time.Duration
		prometheusTimeout string
		expectedResult    time.Duration
	}{
		{
			name:              "higher_prometheus_timeout",
			flagTimeout:       1 * time.Second,
			prometheusTimeout: "10",
			expectedResult:    1 * time.Second,
		},
		{
			name:              "lower_prometheus_timeout",
			flagTimeout:       10 * time.Second,
			prometheusTimeout: "1",
			expectedResult:    1 * time.Second,
		},
		{
			name:              "empty_prometheus_timeout",
			flagTimeout:       1 * time.Second,
			prometheusTimeout: "",
			expectedResult:    1 * time.Second,
		},
		{
			name:              "wrong_prometheus_timeout",
			flagTimeout:       1 * time.Second,
			prometheusTimeout: "not-parsable",
			expectedResult:    1 * time.Second,
		},
	}

	for _, c := range cases {
		request, _ := http.NewRequest("", "", nil)
		request.Header.Set("X-Prometheus-Scrape-Timeout-Seconds", c.prometheusTimeout)
		timeout := getTimeout(request, c.flagTimeout)
		if timeout != c.expectedResult {
			t.Errorf("Test case : %v failed.\nExpected : %v\nGot : %v\n", c.name, c.expectedResult, timeout)
		}
	}
}

func TestCreateLogger(t *testing.T) {
	_, err := createLogger("unexpected")
	if err == nil {
		t.Errorf("logger created with unexpected level")
	}

	for _, lvl := range []string{"error", "warn", "info", "debug"} {
		_, err := createLogger(lvl)
		if err != nil {
			t.Errorf("failed to create logger with level : %v", lvl)
		}
	}
}
