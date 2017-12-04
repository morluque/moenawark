package server

import (
	"database/sql"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model"
	"github.com/morluque/moenawark/sqlstore"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const (
	// MaxBodyLength is the maximum body size in bytes that a client can send us.
	MaxBodyLength = 1024 * 1024
)

type httpError struct {
	code    int
	message string
	err     error
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.message)
}

type resourceMethod0 func(*sql.Tx, http.ResponseWriter, *http.Request) *httpError
type resourceMethod1 func(*sql.Tx, http.ResponseWriter, *http.Request, string) *httpError
type resourceMethodUpdate0 func(*sql.Tx, http.ResponseWriter, *http.Request) *httpError
type resourceMethodUpdate1 func(*sql.Tx, http.ResponseWriter, *http.Request, string) *httpError

type resourceHandler struct {
	listMethod   resourceMethod0
	postMethod   resourceMethodUpdate0
	getMethod    resourceMethod1
	putMethod    resourceMethodUpdate1
	deleteMethod resourceMethodUpdate1
}

type session struct {
	user  *model.User
	since time.Time
}

type apiServerV1 struct {
	TokenLength     int
	TokenHeader     string
	SessionDuration time.Duration
	apiPrefix       string
	apiVersion      string
	sessionLock     sync.RWMutex
	sessionList     map[string]session
	db              *sql.DB
	handlerFuncs    map[string]http.HandlerFunc
}

func newapiServerV1() *apiServerV1 {
	srv := new(apiServerV1)
	srv.apiVersion = "v1"
	srv.TokenLength = config.Cfg.Auth.TokenLength
	srv.TokenHeader = config.Cfg.Auth.TokenHeader
	d, err := time.ParseDuration(config.Cfg.Auth.SessionDuration)
	if err != nil {
		log.Fatalf(
			"Error in config file, could not parse SessionDuration (%s): %s",
			config.Cfg.Auth.SessionDuration,
			err.Error(),
		)
	}
	srv.SessionDuration = d
	srv.sessionLock = sync.RWMutex{}
	srv.sessionList = make(map[string]session)
	srv.handlerFuncs = make(map[string]http.HandlerFunc)

	return srv
}

func (srv *apiServerV1) register(prefix string, h resourceHandler) {
	fullPrefix := fmt.Sprintf("%s/%s/%s", config.Cfg.APIPrefix, srv.apiVersion, prefix)
	reStr := fmt.Sprintf("^%s([^/]+)?$", fullPrefix)
	re, err := regexp.Compile(reStr)
	if err != nil {
		log.Fatal(err)
	}

	srv.handlerFuncs[fullPrefix] = func(w http.ResponseWriter, r *http.Request) {
		subMatches := re.FindStringSubmatch(r.URL.Path)
		if subMatches == nil {
			http.NotFound(w, r)
			return
		}

		// Open DB transaction for create/update/delete
		tx, err := srv.db.BeginTx(r.Context(), nil)
		if err != nil {
			sendError(w, appError(err))
			return
		}
		// The h.*Method() will take care to commit tx if they write to w; else
		// we assume an error occured and we rollback any work.
		defer tx.Rollback()

		var herr *httpError
		log.Printf("debug: method=%s", r.Method)
		switch r.Method {
		case http.MethodGet:
			if len(subMatches[1]) == 0 {
				herr = h.listMethod(tx, w, r)
			} else {
				herr = h.getMethod(tx, w, r, subMatches[1])
			}
		case http.MethodPost:
			herr = h.postMethod(tx, w, r)
		case http.MethodPut:
			herr = h.putMethod(tx, w, r, subMatches[1])
		case http.MethodDelete:
			herr = h.deleteMethod(tx, w, r, subMatches[1])
		default:
			herr = unknownMethodError(r.Method)
		}
		if herr != nil {
			// We are responsible to send the HTTP error to the client
			log.Printf("info: %d %s", herr.code, herr.message)
			sendError(w, herr)
			return
		}
	}
}

// ServeHTTP starts an HTTP server for the JSON REST API
func ServeHTTP() {
	srv1 := newapiServerV1()
	db, err := sqlstore.Open(config.Cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	srv1.db = db

	srv1.register("user", newUserHandler())
	srv1.register("auth", newAuthHandler())

	hmux := http.NewServeMux()
	for prefix, handlerFunc := range srv1.handlerFuncs {
		hmux.HandleFunc(prefix, handlerFunc)
	}

	http.ListenAndServe(config.Cfg.HTTPListen, hmux)
}

func sendError(w http.ResponseWriter, e *httpError) {
	http.Error(w, e.message, e.code)
}

func notFoundError() *httpError {
	return &httpError{code: 404, message: "Resource not found"}
}

func appError(err error) *httpError {
	log.Printf("error: %s", err.Error())
	return &httpError{code: 500, message: "Internal server error"}
}

func userError(err error) *httpError {
	log.Printf("warning: %s", err.Error())
	return &httpError{code: 400, message: "Bad request"}
}

func authError(err error) *httpError {
	log.Printf("auth error: %s", err.Error())
	return &httpError{code: 403, message: "Forbidden"}
}

func unknownMethodError(method string) *httpError {
	return &httpError{code: 405, message: fmt.Sprintf("Method not allowed: %s", method)}
}
