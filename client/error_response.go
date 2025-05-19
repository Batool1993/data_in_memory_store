package client

import "fmt"

// HTTPError represents an error returned by the server.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements error.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}
