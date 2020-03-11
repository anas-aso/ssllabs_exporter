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
	"context"
	"time"

	"github.com/anas-aso/ssllabs_exporter/pkg/ssllabs"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Handle runs SSLLabs assessment on the specified target
// and returns a Prometheus Registry with the results
func Handle(ctx context.Context, logger log.Logger, target string) prometheus.Gatherer {
	var (
		registry           = prometheus.NewRegistry()
		probeDurationGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ssllabs_probe_duration_seconds",
			Help: "Displays how long the assessment took to complete in seconds",
		})
		probeSuccessGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ssllabs_probe_success",
			Help: "Displays whether the assessment succeeded or not",
		})

		probeGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ssllabs_grade",
			Help: "Displays the returned SSLLabs grade of the target host",
		}, []string{"grade"})
		probeAgeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ssllabs_grade_age_seconds",
			Help: "Displays the assessment time for the target host",
		})
	)

	registry.MustRegister(probeDurationGauge)
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeGaugeVec)
	registry.MustRegister(probeAgeGauge)

	start := time.Now()
	result, err := ssllabs.Analyze(ctx, logger, target)

	probeDurationGauge.Set(time.Since(start).Seconds())

	if err != nil {
		level.Error(logger).Log("msg", "assessment failed", "target", target, "error", err)
		// set the probe date to now if the assessment failed
		probeAgeGauge.Set(float64(time.Now().Unix()))
		// set grade to -1 if the assessment failed
		probeGaugeVec.WithLabelValues("-").Set(-1)

		return registry
	}

	probeSuccessGauge.Set(1)

	grade := endpointsLowestGrade(result.Endpoints)

	if grade != "" {
		probeGaugeVec.WithLabelValues(grade).Set(1)

		// TestTime is in milliseconds
		probeAgeGauge.Set(float64(result.TestTime / 1000))
	} else {
		// set grade to 0 if the target does not have an endpoint
		probeGaugeVec.WithLabelValues("-").Set(0)

		// set the probe date to now if the result does not have a grade
		probeAgeGauge.Set(float64(time.Now().Unix()))
	}

	return registry
}
