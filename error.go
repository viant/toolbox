package toolbox

import "io"

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
