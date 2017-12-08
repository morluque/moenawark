package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model"
	"net/http"
	"sync"
	"time"
)

var (
	sessionList     = make(map[string]session)
	sessionLock     = sync.RWMutex{}
	sessionDuration = time.Hour * 2
)

type session struct {
	user  *model.User
	since time.Time
}

func setSessionDuration() {
	str := config.Get("auth.session_duration")
	d, err := time.ParseDuration(str)
	if err != nil {
		log.Errorf("invalid session duration %s, keeping previous value", str)
		return
	}
	sessionLock.Lock()
	defer sessionLock.Unlock()
	sessionDuration = d
}

func isExpiredSession(now time.Time, s session) bool {
	return now.After(s.since.Add(sessionDuration))
}

func createSession(user *model.User) string {
	reapSessions()
	token := createAuthToken()
	s := session{user: user, since: time.Now()}
	sessionLock.Lock()
	defer sessionLock.Unlock()
	sessionList[token] = s

	return token
}

func getSession(token string) (*session, error) {
	sessionLock.RLock()
	defer sessionLock.RUnlock()
	s, ok := sessionList[token]
	if !ok {
		return nil, fmt.Errorf("no such session %s", token)
	}
	return &s, nil
}

func deleteSession(token string) {
	sessionLock.Lock()
	defer sessionLock.Unlock()
	delete(sessionList, token)
	log.Debugf("session %s deleted", token)
}

func getExpiredTokens() []string {
	tokens := make([]string, 0)
	now := time.Now()

	sessionLock.RLock()
	defer sessionLock.RUnlock()
	for t, s := range sessionList {
		if isExpiredSession(now, s) {
			tokens = append(tokens, t)
		}
	}

	return tokens
}

func reapSessions() {
	expiredTokens := getExpiredTokens()
	sessionLock.Lock()
	defer sessionLock.Unlock()
	for _, t := range expiredTokens {
		delete(sessionList, t)
	}
}

func getAuthToken(r *http.Request) (*string, error) {
	tokenHeader, ok := r.Header[config.Get("auth.token_header")]
	log.Debugf("tokenheader=%v", tokenHeader)
	if !ok {
		return nil, fmt.Errorf("missing auth header")
	}
	return &tokenHeader[0], nil
}

func getAuthUser(r *http.Request) (*model.User, error) {
	token, err := getAuthToken(r)
	if err != nil {
		return nil, err
	}
	s, err := getSession(*token)
	if err != nil {
		return nil, err
	}
	if isExpiredSession(time.Now(), *s) {
		return nil, fmt.Errorf("session %s expired", *token)
	}
	return s.user, nil
}

func createAuthToken() string {
	b := make([]byte, config.GetInt("auth.token_length"))
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

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
	token := createSession(user)
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
	token, err := getAuthToken(r)
	if err != nil {
		return authError(err)
	}
	deleteSession(*token)
	return nil
}
