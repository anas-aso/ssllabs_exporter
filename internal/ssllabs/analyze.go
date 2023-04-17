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

package ssllabs

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	ssllabsApi "github.com/essentialkaos/sslscan/v13"
	log "github.com/rs/zerolog"

	"github.com/anas-aso/ssllabs_exporter/internal/build"
)

var api *ssllabsApi.API

func init() {
	api, _ = ssllabsApi.NewAPI("ssllabs-exporter", build.Version)
	if api == nil {
		panic("failed to initialize API client. this should never happen!")
	}
}

// Analyze executes the SSL test HTTP requests
func Analyze(ctx context.Context, logger log.Logger, target string) (result *ssllabsApi.AnalyzeInfo, err error) {
	logger.Debug().Str("target", target).Msg("start processing")

	// check cached results and return them if they are "fresh enough"
	// this is mainly useful if the previous context timed out or
	// canceled before we collected the results
	analyzeProgress, err := api.Analyze(target, ssllabsApi.AnalyzeParams{})
	if err != nil {
		logger.Error().Err(err).Str("target", target).Msg("failed to get cached result")
		return
	}

	result, err = analyzeProgress.Info(true, false)
	if err != nil {
		logger.Error().Err(err).Str("target", target).Msg("failed to get cached result")
		return
	}

	deadline, _ := ctx.Deadline()
	// reconstruct the assessment timeout from the context deadline
	timeout := deadline.Unix() - time.Now().Unix()
	if result.Status == ssllabsApi.STATUS_READY && result.TestTime/1000+timeout >= time.Now().Unix() {
		logger.Debug().Str("target", target).Msg("cached result will be used")
		return
	}

	// trigger a new assessment if there isn't one in progress
	if result.Status != ssllabsApi.STATUS_DNS && result.Status != ssllabsApi.STATUS_IN_PROGRESS {
		logger.Debug().Str("target", target).Msg("triggering a new assessment")
		analyzeProgress, err = api.Analyze(target, ssllabsApi.AnalyzeParams{StartNew: true})
		if err != nil {
			logger.Error().Err(err).Str("target", target).Msg("failed to trigger a new assessment")
			return
		}
	}

	result, err = analyzeProgress.Info(true, false)
	if err != nil {
		logger.Error().Err(err).Str("target", target).Msg("failed to get running assessment info")
		return
	}

	for {
		switch {
		case result.Status == ssllabsApi.STATUS_READY:
			logger.Debug().Str("target", target).Msg("assessment finished successfully")
			return result, nil
		case time.Now().After(deadline):
			result.Status = StatusDeadlineExceeded
			return result, fmt.Errorf("context deadline exceeded")
		// fetch updates at random intervals
		default:
			time.Sleep(time.Duration(10+rand.Intn(10)) * time.Second)
			logger.Debug().Str("target", target).Msg("fetching assessment updates")
			result, err = analyzeProgress.Info(true, false)
			if err != nil {
				logger.Error().Err(err).Str("target", target).Msg("failed to fetch updates")
				return
			}
		}
	}
}
