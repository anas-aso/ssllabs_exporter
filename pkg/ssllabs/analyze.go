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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	log "github.com/rs/zerolog"
)

// Result a lightweight version of the returned
// result from /analyze endpoint. We will parse only the
// parts that we need for the purpose of the exporter
type Result struct {
	// assessment processing status
	// possible values for this field are defined in this package
	// as constants with the name "Status<Type>" (e.g StatusDNS)
	Status string `json:"status"`

	// timestamp when the assessment finished in milliseconds
	TestTime int64 `json:"testTime"`

	// individual target endpoints (IPs) results
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint result of each domain endpoint in Result
type Endpoint struct {
	// endpoint assessment status
	StatusMessage string `json:"statusMessage"`

	// endpoint assessment result
	Grade string `json:"grade"`
}

// Analyze executes the SSL test HTTP requests and
// returns an Result and error (if any)
func Analyze(ctx context.Context, logger log.Logger, target string) (result Result, err error) {
	logger.Debug().Str("target", target).Msg("start processing")

	// check cached results and return them if they are "fresh enough"
	// this is mainly useful if the previous context timed out or
	// canceled before we collected the results
	result, err = analyze(ctx, logger, target, false)
	if err != nil {
		logger.Error().Err(err).Str("target", target).Msg("failed to get cached result")
		return
	}

	deadline, _ := ctx.Deadline()
	// reconstruct the assessment timeout from the context deadline
	timeout := deadline.Unix() - time.Now().Unix()

	if result.Status == StatusReady && result.TestTime/1000+timeout >= time.Now().Unix() {
		logger.Debug().Str("target", target).Msg("cached result will be used")
		return
	}

	// trigger a new assessment if there isn't one in progress
	if result.Status != StatusDNS && result.Status != StatusInProgress {
		logger.Debug().Str("target", target).Msg("triggering a new assessment")
		result, err = analyze(ctx, logger, target, true)
		if err != nil {
			logger.Error().Err(err).Str("target", target).Msg("failed to trigger a new assessment")
			return
		}
	}

	for {
		switch {
		case result.Status == StatusReady:
			logger.Debug().Str("target", target).Msg("assessment finished successfully")
			return result, nil
		case time.Now().After(deadline):
			result.Status = StatusDeadlineExceeded
			return result, fmt.Errorf("context deadline exceeded")
		// fetch updates at random intervals
		default:
			time.Sleep(time.Duration(10+rand.Intn(10)) * time.Second)
			logger.Debug().Str("target", target).Msg("fetching assessment updates")
			result, err = analyze(ctx, logger, target, false)
			if err != nil {
				logger.Error().Err(err).Str("target", target).Msg("failed to fetch updates")
				return
			}
		}
	}
}

// retry API calls until we get a 200 response or the deadline is reached
// this function is intended to take care of auto retrying when facing network
// failures, remote server failures and/or rate limiting.
func analyze(ctx context.Context, logger log.Logger, target string, new bool) (Result, error) {
	var result Result
	deadline, _ := ctx.Deadline()
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			result.Status = StatusAborted
			return result, fmt.Errorf("context canceled")
		default:
			switch result = getAnalyze(target, new); {
			case result.Status == StatusDNS || result.Status == StatusInProgress || result.Status == StatusReady:
				return result, nil
			case result.Status == StatusError:
				return result, fmt.Errorf("the remote server couldn't process the request")
			case result.Status == StatusHTTPError:
				coolOff := time.Duration(rand.Intn(10)) * time.Second
				logger.Debug().Str("target", target).Dur("duration", coolOff).Msg("sleeping due to HTTP error")
				time.Sleep(coolOff)
			case result.Status == StatusServerError:
				coolOff := time.Duration(30+rand.Intn(30)) * time.Second
				logger.Debug().Str("target", target).Dur("duration", coolOff).Msg("sleeping due to remote server error")
				time.Sleep(coolOff)
			default:
				return result, fmt.Errorf("unrecognized status: %v", result.Status)
			}
		}
		// always reset the result by the end of every iteration
		result = Result{}
	}

	result.Status = StatusDeadlineExceeded
	return result, fmt.Errorf("context deadline exceeded")
}

// invokes SSLLabs API /analyze endpoint and encapsulate the result in an Result
func getAnalyze(target string, new bool) (result Result) {
	request := API + "analyze?host=" + target + "&all=done"
	if new {
		request += "&startNew=on"
	}

	// TODO: make http timeout configurable
	httpClient := http.Client{Timeout: 1 * time.Minute}
	response, err := httpClient.Get(request)
	if err != nil {
		result.Status = StatusHTTPError
		return
	}

	defer response.Body.Close()

	// this should happen in case of 429 or 5xx errors
	if response.StatusCode != http.StatusOK {
		result.Status = StatusServerError
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		result.Status = StatusHTTPError
		return
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		result.Status = StatusHTTPError
	}

	return
}
