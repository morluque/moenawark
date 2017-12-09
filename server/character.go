package server

import (
	"database/sql"
	"net/http"
)

// CharacterHandler is a resource handler for in-game characters.
type CharacterHandler struct {
	*resourceMapper
}

// SetResourceMapper sets the resourceMapper that can be used to create URLs to arbitrary resources.
func (h CharacterHandler) SetResourceMapper(m *resourceMapper) {
	h.resourceMapper = m
}

// View reponds with JSON representing an in-game character controlled by a user.
func (h CharacterHandler) View(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
}

// List will list all characters as JSON.
func (h CharacterHandler) List(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	return unknownMethodError(r.Method)
}

// Create adds a new character to the database.
func (h CharacterHandler) Create(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	return unknownMethodError(r.Method)
}

// Update modifies a character from the user-supplied JSON body.
func (h CharacterHandler) Update(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
}

// Delete would delete a character from database if it was implemented.
func (h CharacterHandler) Delete(db *sql.Tx, w http.ResponseWriter, r *http.Request, unused string) *httpError {
	return unknownMethodError(r.Method)
}
