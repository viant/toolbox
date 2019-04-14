package toolbox

import (
	"fmt"
	"io"
	"strings"
)

//NilPointerError represents nil pointer error
type NilPointerError struct {
	message string
}

//Error returns en error
func (e *NilPointerError) Error() string {
	if e.message == "" {
		return "NilPointerError"
	}
	return e.message
}

//NewNilPointerError creates a new nil pointer error
func NewNilPointerError(message string) error {
	return &NilPointerError{
		message: message,
	}
}

//IsNilPointerError returns true if error is nil pointer
func IsNilPointerError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NilPointerError)
	return ok
}

//IsEOFError returns true if io.EOF
func IsEOFError(err error) bool {
	if err == nil {
		return false
	}
	return err == io.EOF
}

//NotFoundError represents not found error
type NotFoundError struct {
	URL string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %v", e.URL)
}

//IsNotFoundError checks is supplied error is NotFoundError type
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}

//ReclassifyNotFoundIfMatched reclassify error if not found
func ReclassifyNotFoundIfMatched(err error, URL string) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "doesn't exist") ||
		strings.Contains(message, "no such file or directory") ||
		strings.Contains(err.Error(), "404") ||
		strings.Contains(err.Error(), "nosuchbucket") {
		return &NotFoundError{URL: URL}
	}
	return err
}
