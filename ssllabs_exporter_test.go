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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestProbeHandler(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
	}))
	defer testServer.Close()

	req, err := http.NewRequest("GET", "?target="+testServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Prometheus-Scrape-Timeout-Seconds", "1")

	testRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, log.NewNopLogger(), 1, newCache(1, 1))
	})

	handler.ServeHTTP(testRecorder, req)

	if status := testRecorder.Code; status != http.StatusOK {
		t.Errorf("probe handler returned the wrong status code.\nExpected : %v\nGot : %v\n", status, http.StatusOK)
	}
}

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

func TestValidateTimeout(t *testing.T) {
	var cases = []struct {
		name            string
		flagTimeout     string
		expectedTimeout time.Duration
		expectedError   error
	}{
		{
			name:            "un_parsable_duration",
			flagTimeout:     "not_a_duration",
			expectedTimeout: 0,
			expectedError:   errors.New(`time: invalid duration "not_a_duration"`),
		},
		{
			name:            "less_than_1m",
			flagTimeout:     "1s",
			expectedTimeout: 0,
			expectedError:   errors.New("probe timeout must be a least 1 minute"),
		},
		{
			name:            "good_value",
			flagTimeout:     "5m",
			expectedTimeout: 5 * time.Minute,
			expectedError:   nil,
		},
	}

	for _, c := range cases {
		timeout, err := validateTimeout(c.flagTimeout)
		if err != nil && err.Error() != c.expectedError.Error() || err == nil && c.expectedError != nil || timeout != c.expectedTimeout {
			t.Errorf("Test case : %v failed.\nExpected : (%v, %v)\nGot : (%v, %v)\n", c.name, c.expectedTimeout, c.expectedError, timeout, err)
		}
	}
}
