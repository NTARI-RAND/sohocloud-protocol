package httpjson

// HTTPStatusError is returned when the coordinator responds with a non-2xx
// HTTP status. Its diagnostic includes at most 4096 bytes of the response body.
type HTTPStatusError struct {
	statusCode int
	message    string
}

func (e *HTTPStatusError) Error() string { return e.message }

// StatusCode returns the coordinator's HTTP response status code.
func (e *HTTPStatusError) StatusCode() int { return e.statusCode }
