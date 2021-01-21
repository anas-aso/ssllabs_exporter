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

// Modifications copyright (C) 2021 diamonwiggins

package ssllabs

var gradesMapping = map[string]float64{
	"A":     (80 + 100) / 2,
	"A+":    ((80 + 100) / 2) + 1,
	"A-":    ((80 + 100) / 2) - 1,
	"B":     (65 + 80) / 2,
	"C":     (50 + 65) / 2,
	"D":     (35 + 50) / 2,
	"E":     (20 + 35) / 2,
	"F":     (0 + 20) / 2,
	"M":     0,
	"T":     0,
	"undef": -1,
}

// convert the returned grade to a number based on https://github.com/ssllabs/research/wiki/SSL-Server-Rating-Guide
func EndpointsLowestGrade(ep []Endpoint) (result string) {
	if len(ep) == 0 {
		return
	}

	// the target host gets the lowest score of its endpoints
	for _, e := range ep {
		// skip endpoints without a grade : case of unreachable endpoint(s)
		if e.Grade == "" {
			continue
		}

		// initialize the result with the first defined grade
		if result == "" {
			result = e.Grade
		}

		eGrade, ok := gradesMapping[e.Grade]
		if ok {
			if gradesMapping[result] > eGrade {
				result = e.Grade
			}
		} else {
			return "undef"
		}
	}

	return
}
