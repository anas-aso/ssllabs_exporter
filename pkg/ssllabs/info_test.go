package ssllabs

import (
	"reflect"
	"testing"
)

func TestInfo(t *testing.T) {
	info, err := Info()
	if err != nil {
		t.Errorf("\nerror fetching SSLLabs API info: %v", err)
	}

	expectedInfo := APIInfo{
		EngineVersion:      "2.1.0",
		CriteriaVersion:    "2009q",
		MaxAssessments:     25,
		CurrentAssessments: 0,
	}

	if !reflect.DeepEqual(expectedInfo, info) {
		t.Errorf("\nexpected info: %v\nreturned info: %v", expectedInfo, info)
	}
}
