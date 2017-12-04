package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/morluque/moenawark/model"
	"net/http"
	"sync"
	"time"
)

const (
	// TokenLength is the length in bytes of the security token
	TokenLength = 32
	// TokenHeader is the name of the header used to communicate the security token
	TokenHeader = "X-Auth-Token"
)

var (
	sessionList     = make(map[string]session)
	sessionLock     = sync.RWMutex{}
	sessionDuration = time.Hour * 2
)

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
	tokenHeader, ok := r.Header[TokenHeader]
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
	b := make([]byte, TokenLength)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func authGet(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	log.Debugf("authGet got called")
	return unknownMethodError(r.Method)
}

func authList(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	return unknownMethodError(r.Method)
}

func authCreate(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	user, err := model.AuthUser(db, login, password)
	if err != nil {
		return authError(err)
	}
	token := createSession(user)
	headers := w.Header()
	headers[TokenHeader] = []string{token}
	log.Infof("user %s successfully logged in", login)
	return nil
}

func authUpdate(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
}

func authDelete(db *sql.Tx, w http.ResponseWriter, r *http.Request, unused string) *httpError {
	token, err := getAuthToken(r)
	if err != nil {
		return authError(err)
	}
	deleteSession(*token)
	return nil
}

func newAuthHandler() resourceHandler {
	return resourceHandler{
		getMethod:    authGet,
		listMethod:   authList,
		putMethod:    authUpdate,
		postMethod:   authCreate,
		deleteMethod: authDelete,
	}
}
