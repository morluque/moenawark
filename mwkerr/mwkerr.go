/*
Package mwkerr defines a custom error type with numerical code for Moenawark.

Maybe this wasn't such a good idea after all...
*/
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
	// DuplicateModel signals that a model object already exists in database
	DuplicateModel
	// AuthError signals an Email/password mismatch
	AuthError
	// DatabaseEmpty signals that the db must be initialized
	DatabaseEmpty
	// DatabaseAlreadyInitialized signals that you can't init an existing database
	DatabaseAlreadyInitialized
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
