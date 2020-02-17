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
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/anas-aso/ssllabs_exporter/pkg/exporter"
)

var (
	listenAddress = kingpin.Flag("listen-address", "The address to listen on for HTTP requests.").Default(":19115").String()
	probeTimeout  = kingpin.Flag("timeout", "Assessment timeout in seconds (including retries).").Default("600").Float64()
	logLevel      = kingpin.Flag("log-level", "Printed logs level.").Default("debug").Enum("error", "warn", "info", "debug")
	version       string
)

func probeHandler(w http.ResponseWriter, r *http.Request, logger log.Logger, timeoutSeconds float64) {
	params := r.URL.Query()
	target := params.Get("target")
	if target == "" {
		// TODO: add more validation for the target (e.g valid hostname, DNS, etc)
		level.Error(logger).Log("msg", "Target parameter is missing")
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	if t, ok := scrapeTimeoutExists(r); ok {
		timeoutSeconds = t * float64(time.Second)
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds))
	defer cancel()
	r = r.WithContext(ctx)

	registry := exporter.Handle(ctx, logger, target)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	var logger log.Logger
	var lvl level.Option
	switch *logLevel {
	case "error":
		lvl = level.AllowError()
	case "warn":
		lvl = level.AllowWarn()
	case "info":
		lvl = level.AllowInfo()
	case "debug":
		lvl = level.AllowDebug()
	default:
		panic("unexpected log level")
	}
	logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, lvl)
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)

	timeoutSeconds := *probeTimeout * float64(time.Second)
	// A new assessment will always take at least 60 seconds per host
	// endpoint. A timeout less than 60 seconds doesn't make sense.
	if timeoutSeconds < float64(60*time.Second) {
		level.Warn(logger).Log("msg", "configured timeout is less than 60 seconds. switching to default timeout")
		timeoutSeconds = float64(600 * time.Second)
	}

	level.Info(logger).Log("msg", "Starting ssllabs_exporter", "version", version)

	// TODO: expose SSLLabs API info (ssllabs.Info())
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, logger, timeoutSeconds)
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

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

// get scrape timeout value if defined
func scrapeTimeoutExists(r *http.Request) (float64, bool) {
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		timeoutSeconds, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return timeoutSeconds, true
	}
	return 0, false
}
