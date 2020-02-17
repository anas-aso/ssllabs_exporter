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

const (
	// API SSLLabs API URL
	API = "https://api.ssllabs.com/api/v3/"
	// StatusDNS assessment still in DNS resolution phase
	StatusDNS = "DNS"
	// StatusError error running the assessment (e.g target behind a firewall)
	StatusError = "ERROR"
	// StatusInProgress assessment in progress
	StatusInProgress = "IN_PROGRESS"
	// StatusReady SSLLabs assessment finished successfully
	StatusReady = "READY"
	// StatusHTTPError error processing the HTTP request to SSLLabs API
	StatusHTTPError = "HTTP_ERROR"
	// StatusServerError SSLLabs API server error or rate limiting
	StatusServerError = "SERVER_ERROR"
	// StatusDeadlineExceeded assessment deadline exceeded
	StatusDeadlineExceeded = "DEADLINE_EXCEEDED"
	// StatusAborted assessment canceled by the client
	StatusAborted = "ABORTED"
)
