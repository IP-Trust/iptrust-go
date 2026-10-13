package iptrust

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrMissingAPIKey is returned by NewClient when the API key is empty.
var ErrMissingAPIKey = errors.New("iptrust: an API key is required")

// Error is returned when the IP Trust API responds with a non-2xx status.
//
// Use errors.As to inspect it:
//
//	var apiErr *iptrust.Error
//	if errors.As(err, &apiErr) {
//		switch apiErr.StatusCode {
//		case http.StatusTooManyRequests: // quota or rate limit
//		case http.StatusUnauthorized:    // bad key
//		}
//	}
type Error struct {
	// StatusCode is the HTTP status of the failed response.
	StatusCode int
	// Code is the machine-readable error code from the API, e.g.
	// "token_invalid". Empty if the body was not a recognised error.
	Code string
	// Detail is a human-friendly description from the API, if present.
	Detail string
	// ErrorID is a correlation ID set on internal errors (5xx). Quote it
	// when contacting support.
	ErrorID string
	// Data is extra structured data attached to the error, if present.
	Data json.RawMessage
	// Body is the raw response body, truncated, for errors the API did not
	// describe in its usual JSON format.
	Body string
}

// Error implements the error interface.
func (e *Error) Error() string {
	summary := e.Detail
	if summary == "" {
		summary = e.Code
	}
	if summary == "" {
		summary = e.Body
	}
	if summary == "" {
		summary = http.StatusText(e.StatusCode)
	}
	if summary == "" {
		return fmt.Sprintf("iptrust: request failed with status %d", e.StatusCode)
	}
	return fmt.Sprintf("iptrust: request failed with status %d: %s", e.StatusCode, summary)
}

const maxErrorBody = 4096

// newErrorFromResponse builds an *Error from a failed HTTP response. It
// reads (but does not close) the body.
func newErrorFromResponse(resp *http.Response) *Error {
	e := &Error{StatusCode: resp.StatusCode}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))

	var body apiErrorBody
	if err := json.Unmarshal(raw, &body); err == nil && body.Error.ErrorCode != "" {
		e.Code = body.Error.ErrorCode
		e.Detail = body.Error.Detail
		e.ErrorID = body.Error.ErrorID
		e.Data = body.Error.Data
		return e
	}

	e.Body = strings.TrimSpace(string(raw))
	if len(e.Body) > 200 {
		e.Body = e.Body[:200]
	}
	return e
}
