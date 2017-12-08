package server

import (
	"database/sql"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model"
	"github.com/morluque/moenawark/server/session"
	"net/http"
)

// AuthHandler is a resource handler for user authentication.
type AuthHandler struct {
	*resourceMapper
}

// SetResourceMapper sets the resourceMapper that can be used to create URLs to arbitrary resources.
func (h AuthHandler) SetResourceMapper(m *resourceMapper) {
	h.resourceMapper = m
}

// View handles HTTP GET on an authenticated session (unimplemented).
func (h AuthHandler) View(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	log.Debugf("authGet got called")
	return unknownMethodError(r.Method)
}

// List handles HTTP GET on a collection of authenticated sessions (unimplemented).
func (h AuthHandler) List(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	return unknownMethodError(r.Method)
}

// Create verifies user credentials on HTTP POST and returns a security token.
func (h AuthHandler) Create(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	user, err := model.AuthUser(db, login, password)
	if err != nil {
		return authError(err)
	}
	token := session.Create(user)
	headers := w.Header()
	headers[config.Get("auth.token_header")] = []string{token}
	log.Infof("user %s successfully logged in", login)
	return nil
}

// Update handles HTTP PUT on an authenticated session (unimplemented).
func (h AuthHandler) Update(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
}

// Delete logs out a user, forgetting about it's authentication token.
func (h AuthHandler) Delete(db *sql.Tx, w http.ResponseWriter, r *http.Request, unused string) *httpError {
	err := session.Delete(r)
	if err != nil {
		return authError(err)
	}
	return nil
}
