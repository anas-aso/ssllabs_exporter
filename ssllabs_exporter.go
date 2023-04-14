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
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/rs/zerolog"

	"github.com/anas-aso/ssllabs_exporter/internal/build"
	"github.com/anas-aso/ssllabs_exporter/internal/exporter"
	"github.com/anas-aso/ssllabs_exporter/internal/ssllabs"
)

const (
	pruneDelay = 1 * time.Minute
)

var (
	listenAddress     = kingpin.Flag("listen-address", "The address to listen on for HTTP requests.").Default(":19115").String()
	probeTimeout      = kingpin.Flag("timeout", "Time duration before canceling an ongoing probe such as 30m or 1h5m. This value must be at least 1m. Valid duration units are ns, us (or µs), ms, s, m, h.").Default("10m").String()
	logLevel          = kingpin.Flag("log-level", "Printed logs level.").Default("debug").Enum("error", "warn", "info", "debug")
	cacheRetention    = kingpin.Flag("cache-retention", "Time duration to keep entries in cache such as 30m or 1h5m. Valid duration units are ns, us (or µs), ms, s, m, h.").Default("1h").String()
	cacheIgnoreFailed = kingpin.Flag("cache-ignore-failed", "Do not cache failed results due to intermittent SSLLabs issues.").Default("False").Bool()
)

func probeHandler(w http.ResponseWriter, r *http.Request, logger log.Logger, timeoutSeconds time.Duration, resultsCache *cache) {
	target := r.URL.Query().Get("target")
	// TODO: add more validation for the target (e.g valid hostname, DNS, etc)
	if target == "" {
		logger.Error().Msg("Target parameter is missing")
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	// check if the results are available in the cache
	registry := resultsCache.get(target)

	if registry != nil {
		logger.Debug().Str("target", target).Msg("serving results from cache")
	} else {
		// if the results do not exist in the cache, trigger a new assessment

		timeoutSeconds = getTimeout(r, timeoutSeconds)

		ctx, cancel := context.WithTimeout(r.Context(), timeoutSeconds)
		defer cancel()

		r = r.WithContext(ctx)

		registry = exporter.Handle(ctx, logger, target)

		// do not cache failed assessments if configured
		if *cacheIgnoreFailed && exporter.Failed(registry) {
			return
		}

		// add the assessment results to the cache
		resultsCache.add(target, registry)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	kingpin.Version(build.Version)
	kingpin.Parse()

	logger, err := createLogger(*logLevel)
	if err != nil {
		fmt.Printf("failed to create logger with error: %v", err)
		os.Exit(1)
	}

	timeoutSeconds, err := validateTimeout(*probeTimeout)
	if err != nil {
		logger.Error().Err(err).Msg("failed to validate the probe timeout value")
		os.Exit(1)
	}

	cacheRetentionDuration, err := time.ParseDuration(*cacheRetention)
	if err != nil {
		logger.Error().Err(err).Msg("failed to parse the cache retention value")
		os.Exit(1)
	}
	resultsCache := newCache(pruneDelay, cacheRetentionDuration)

	logger.Info().Str("version", build.Version).Msg("Starting ssllabs_exporter")

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "ssllabs_exporter",
			Help: "SSLLabs exporter build parameters",
			ConstLabels: prometheus.Labels{
				"branch":    build.Branch,
				"goversion": build.GoVersion,
				"revision":  build.Revision,
				"version":   build.Version,
			},
		},
		func() float64 { return 1 },
	)

	ssllabsInfo, err := ssllabs.Info()
	if err != nil {
		logger.Error().Err(err).Msg("Could not fetch SSLLabs API Info")
		os.Exit(1)
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "ssllabs_api",
			Help: "SSLLabs API engine and criteria versions",
			ConstLabels: prometheus.Labels{
				"engine":   ssllabsInfo.EngineVersion,
				"criteria": ssllabsInfo.EngineVersion,
			},
		},
		func() float64 { return 1 },
	)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, logger, timeoutSeconds, resultsCache)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
    <head><title>SSLLabs Exporter</title></head>
    <body>
    <h1>SSLLabs Exporter</h1>
    <p><a href="probe?target=prometheus.io">Check SSLLabs grade for prometheus.io</a></p>
    <p><a href="metrics">Exporter Metrics</a></p>
    </body>
    </html>`))
	})

	logger.Info().Str("address", *listenAddress).Msg("Listening on address")

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		logger.Error().Err(err).Msg("Error starting HTTP server")
		os.Exit(1)
	}
}

// get the min of Prometheus scrape timeout (if found) and the flag timeout
func getTimeout(r *http.Request, timeout time.Duration) time.Duration {
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		scrapeTimeout, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return timeout
		}

		scrapeTimeoutSeconds := time.Duration(scrapeTimeout) * time.Second

		if scrapeTimeoutSeconds < timeout {
			return scrapeTimeoutSeconds
		}
	}

	return timeout
}

// create logger with the provided log level
func createLogger(l string) (logger log.Logger, err error) {
	var lvl log.Level
	switch l {
	case "error":
		lvl = log.ErrorLevel
	case "warn":
		lvl = log.WarnLevel
	case "info":
		lvl = log.InfoLevel
	case "debug":
		lvl = log.DebugLevel
	default:
		return log.Nop(), fmt.Errorf("unrecognized log level: %v", l)
	}

	log.MessageFieldName = "msg"
	log.TimestampFieldName = "timestamp"
	log.TimeFieldFormat = time.RFC3339Nano

	log.SetGlobalLevel(lvl)

	logger = log.New(os.Stdout).With().Timestamp().Logger()

	return logger, nil
}

// validate the provided probe timeout
func validateTimeout(timeout string) (time.Duration, error) {
	timeoutSeconds, err := time.ParseDuration(timeout)
	if err != nil {
		return 0, err
	}

	// A new assessment will always take at least 60 seconds per host
	// endpoint. A timeout less than 60 seconds doesn't make sense.
	if timeoutSeconds < time.Minute {
		return 0, errors.New("probe timeout must be a least 1 minute")
	}

	return timeoutSeconds, nil
}
