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

	"github.com/anas-aso/ssllabs_exporter/pkg/ssllabs"
)

func TestEndpointsLowestGradeValue(t *testing.T) {
	var cases = []struct {
		name           string
		data           []ssllabs.Endpoint
		expectedResult int
	}{
		{
			name:           "result_without_endpoints",
			data:           []ssllabs.Endpoint{},
			expectedResult: 0,
		},
		{
			name: "single_grade",
			data: []ssllabs.Endpoint{
				{
					Grade: "A+",
				},
			},
			expectedResult: gradesMapping["A+"],
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
			expectedResult: gradesMapping["B"],
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
			expectedResult: -1,
		},
	}

	for _, c := range cases {
		grade := endpointsLowestGradeValue(c.data)
		if grade != c.expectedResult {
			t.Errorf("Test case : %v failed.\nExpected : %v\nGot : %v\n", c.name, c.expectedResult, grade)
		}
	}
}
