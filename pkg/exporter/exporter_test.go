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

func TestAggregateGrade(t *testing.T) {
	var cases = []struct {
		name           string
		data           ssllabs.Result
		expectedResult float64
	}{
		{
			name:           "result_without_endpoints",
			data:           ssllabs.Result{},
			expectedResult: 0,
		},
		{
			name: "a_plus_grade",
			data: ssllabs.Result{
				Endpoints: []ssllabs.Endpoint{
					{
						Grade: "A+",
					},
				},
			},
			expectedResult: 1,
		},
		{
			name: "a_grade",
			data: ssllabs.Result{
				Endpoints: []ssllabs.Endpoint{
					{
						Grade: "A",
					},
				},
			},
			expectedResult: 1,
		},
		{
			name: "non_a_grade",
			data: ssllabs.Result{
				Endpoints: []ssllabs.Endpoint{
					{
						Grade: "B",
					},
				},
			},
			expectedResult: 0,
		},
		{
			name: "single_failing_endpoint",
			data: ssllabs.Result{
				Endpoints: []ssllabs.Endpoint{
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
			},
			expectedResult: 0,
		},
	}

	for _, c := range cases {
		grade := aggregateGrade(c.data.Endpoints)
		if grade != c.expectedResult {
			t.Errorf("Test case : %v failed.\nExpected : %v\nGot : %v\n", c.name, c.expectedResult, grade)
		}
	}
}
