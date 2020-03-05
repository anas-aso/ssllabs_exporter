package exporter

import (
	"github.com/anas-aso/ssllabs_exporter/pkg/ssllabs"
)

// Map grades to numerical values as defined in https://github.com/ssllabs/research/wiki/SSL-Server-Rating-Guide#methodology-overview
// Since the documented mapping provides a range of values for each grade instead of fixed ones, we take half the documented interval
// to allow mapping case like A+ and A-.
// A special undocumented cases are T and M for which we assign 0.
var gradesMapping = map[string]int{
	"A":  (80 + 100) / 2,
	"A+": ((80 + 100) / 2) + 1,
	"A-": ((80 + 100) / 2) - 1,
	"B":  (65 + 80) / 2,
	"C":  (50 + 65) / 2,
	"D":  (35 + 50) / 2,
	"E":  (20 + 35) / 2,
	"F":  (0 + 20) / 2,
	"M":  0,
	"T":  0,
}

// convert the returned grade to a number based on https://github.com/ssllabs/research/wiki/SSL-Server-Rating-Guide
func endpointsLowestGradeValue(ep []ssllabs.Endpoint) (grade int) {
	if len(ep) == 0 {
		return 0
	}
	// initialize the returned grade to the highest possible value
	grade = 100

	// the target host gets the lowest score of its endpoints
	for _, e := range ep {
		eGrade, ok := gradesMapping[e.Grade]
		if ok {
			if grade > eGrade {
				grade = eGrade
			}
		} else {
			return -1
		}
	}

	return
}
