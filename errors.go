package circleci

import "fmt"

// Error represents an error returned from the circleci.com API.
type Error struct {
	URL        string
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("circleci: %v %v '%.100s'", e.StatusCode, e.URL, e.Body)
}
