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

	"github.com/anas-aso/ssllabs_exporter/internal/ssllabs"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/rs/zerolog"
)

const probeSuccessMetricName = "ssllabs_probe_success"

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
			Name: probeSuccessMetricName,
			Help: "Displays whether the assessment succeeded or not",
		})

		probeGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ssllabs_grade",
			Help: "Displays the returned SSLLabs grade of the target host",
		}, []string{"grade"})
		probeTimeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ssllabs_grade_time_seconds",
			Help: "Displays the assessment time for the target host",
		})
	)

	registry.MustRegister(probeDurationGauge)
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeGaugeVec)
	registry.MustRegister(probeTimeGauge)

	start := time.Now()
	probeTimeGauge.Set(float64(start.Unix()))

	result, err := ssllabs.Analyze(ctx, logger, target)

	probeDurationGauge.Set(time.Since(start).Seconds())

	if err != nil {
		logger.Error().Err(err).Str("target", target).Msg("assessment failed")
		// set grade to -1 if the assessment failed
		probeGaugeVec.WithLabelValues("-").Set(-1)

		return registry
	}

	probeSuccessGauge.Set(1)

	grade := endpointsLowestGrade(result.Endpoints)

	if grade != "" {
		probeGaugeVec.WithLabelValues(grade).Set(1)
	} else {
		// set grade to 0 if the target does not have an endpoint
		probeGaugeVec.WithLabelValues("-").Set(0)
	}

	return registry
}

// Failed checks whether the assessment failed or not
func Failed(registry prometheus.Gatherer) bool {
	metrics, err := registry.Gather()
	if err != nil {
		return false
	}

	for _, m := range metrics {
		if m.GetName() == probeSuccessMetricName {
			result := m.GetMetric()[0].GetGauge().Value
			return *result == 0
		}
	}

	return false
}
