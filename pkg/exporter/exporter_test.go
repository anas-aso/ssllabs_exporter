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

	"github.com/prometheus/client_golang/prometheus"
)

func TestFailed(t *testing.T) {
	var cases = []struct {
		name           string
		value          float64
		expectedResult bool
	}{
		{
			name:           "failed_assessment",
			value:          0,
			expectedResult: true,
		},
		{
			name:           "successful_assessment",
			value:          1,
			expectedResult: false,
		},
		{
			name:           "unregistered_metric",
			value:          -1,
			expectedResult: false,
		},
	}

	for _, c := range cases {
		// init registry
		registry := prometheus.NewRegistry()
		probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: probeSuccessMetricName,
			Help: "Displays whether the assessment succeeded or not",
		})

		if c.value != -1 {
			registry.MustRegister(probeSuccessGauge)
			probeSuccessGauge.Set(c.value)
		}

		result := Failed(registry)
		if result != c.expectedResult {
			t.Errorf("Test case : %v failed.\nExpected : %v\nGot : %v\n", c.name, c.expectedResult, result)
		}
	}
}
