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

package exporter

import (
	"testing"

	"github.com/anas-aso/ssllabs_exporter/internal/ssllabs"
)

func TestEndpointsLowestGrade(t *testing.T) {
	var cases = []struct {
		name           string
		data           []ssllabs.Endpoint
		expectedResult string
	}{
		{
			name:           "result_without_endpoints",
			data:           []ssllabs.Endpoint{},
			expectedResult: "",
		},
		{
			name: "result_with_unreachable_endpoint(s)",
			data: []ssllabs.Endpoint{
				{
					StatusMessage: "Unable to connect to the server",
					Grade:         "",
				},
			},
			expectedResult: "",
		},
		{
			name: "result_with_single_unreachable_endpoint",
			data: []ssllabs.Endpoint{
				{
					StatusMessage: "Unable to connect to the server",
					Grade:         "",
				},
				{
					Grade: "A",
				},
				{
					Grade: "B",
				},
			},
			expectedResult: "B",
		},
		{
			name: "single_grade",
			data: []ssllabs.Endpoint{
				{
					Grade: "A+",
				},
			},
			expectedResult: "A+",
		},
		{
			name: "multiple_grades",
			data: []ssllabs.Endpoint{
				{
					Grade: "B",
				},
				{
					Grade: "A",
				},
				{
					Grade: "A+",
				},
			},
			expectedResult: "B",
		},
		{
			name: "unknown_grade",
			data: []ssllabs.Endpoint{
				{
					Grade: "B",
				},
				{
					Grade: "A",
				},
				{
					Grade: "X",
				},
			},
			expectedResult: "undef",
		},
	}

	for _, c := range cases {
		grade := endpointsLowestGrade(c.data)
		if grade != c.expectedResult {
			t.Errorf("Test case : %v failed.\nExpected : %v\nGot : %v\n", c.name, c.expectedResult, grade)
		}
	}
}
