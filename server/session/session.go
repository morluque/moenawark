package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"github.com/morluque/moenawark/model"
	"net/http"
	"sync"
	"time"
)

var (
	log             *loglevel.Logger
	sessionList     = make(map[string]session)
	sessionLock     = sync.RWMutex{}
	sessionDuration = time.Hour * 2
)

func init() {
	log = loglevel.New("session", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
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

type session struct {
	user  *model.User
	since time.Time
}

// Create associates a user with a security token, and returns that new token.
func Create(user *model.User) string {
	reapSessions()
	token := createAuthToken()
	s := session{user: user, since: time.Now()}
	sessionLock.Lock()
	defer sessionLock.Unlock()
	sessionList[token] = s

	return token
}

// Delete forgets about a user/token pair.
//
// The user will not be considered authenticated any more.
func Delete(r *http.Request) error {
	token, err := getAuthToken(r)
	if err != nil {
		return err
	}
	sessionLock.Lock()
	defer sessionLock.Unlock()
	delete(sessionList, *token)
	log.Debugf("session %s deleted", token)
	return nil
}

// User returns the authenticated user for this request, if any.
func User(r *http.Request) (*model.User, error) {
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

func isExpiredSession(now time.Time, s session) bool {
	return now.After(s.since.Add(sessionDuration))
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

func createAuthToken() string {
	b := make([]byte, config.GetInt("auth.token_length"))
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
