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
