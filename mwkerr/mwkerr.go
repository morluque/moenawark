package mwkerr

import "fmt"

// MWKError is a game-specific error grouping a numeric code with a message.
type MWKError struct {
	Code    int
	Message string
}

const (
	// Unknown error, should not be used
	Unknown = iota
	// DuplicateCharacter signals that a character by this name already exists in database
	DuplicateCharacter
	// DuplicateUser signals that a user with this email already exists in database
	DuplicateUser
	// AuthError signals an Email/password mismatch
	AuthError
)

/*
New creates a MWKError with the given code and message.
It can format the message with SPrintf.
*/
func New(code int, format string, args ...interface{}) MWKError {
	message := fmt.Sprintf(format, args...)
	return MWKError{Code: code, Message: message}
}

func (e MWKError) Error() string {
	return e.Message
}
