package envelope

import "time"

// IngestResult is the answer to POST /v1/events. A batch is never all-or-nothing on content:
// a line that fails validation is listed in Rejected and the rest are stored.
type IngestResult struct {
	Accepted   int         `json:"accepted"`
	Duplicates int         `json:"duplicates"`
	Rejected   []Rejection `json:"rejected,omitempty"`
	ServerTime time.Time   `json:"server_time"`
}

// Rejection names a line (1-based) of the request body and why it was refused.
type Rejection struct {
	Line  int    `json:"line"`
	Field string `json:"field,omitempty"`
	Error string `json:"error"`
}
