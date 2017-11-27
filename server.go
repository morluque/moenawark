package main

import (
	"database/sql"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/sqlstore"
	"log"
	"net/http"
	"regexp"
)

type resourceMethod1 func(*sql.DB, http.ResponseWriter, *http.Request, string)
type resourceMethodUpdate1 func(*sql.Tx, http.ResponseWriter, *http.Request, string)
type resourceMethod0 func(*sql.DB, http.ResponseWriter, *http.Request)
type resourceMethodUpdate0 func(*sql.Tx, http.ResponseWriter, *http.Request)

type resourceHandler struct {
	listMethod   resourceMethod0
	postMethod   resourceMethodUpdate0
	getMethod    resourceMethod1
	putMethod    resourceMethodUpdate1
	deleteMethod resourceMethodUpdate1
}

func (h resourceHandler) register(m *http.ServeMux, db *sql.DB, prefix string) {
	reStr := fmt.Sprintf("^%s([^/]+)?$", prefix)
	re, err := regexp.Compile(reStr)
	if err != nil {
		log.Fatal(err)
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		subMatches := re.FindStringSubmatch(r.URL.Path)
		if subMatches == nil {
			http.NotFound(w, r)
			return
		}

		// Handle read-only actions
		if r.Method == http.MethodGet {
			if len(subMatches[1]) == 0 {
				h.listMethod(db, w, r)
			} else {
				h.getMethod(db, w, r, subMatches[1])
			}
			return
		}

		// Open DB transaction for create/update/delete
		tx, err := db.Begin()
		if err != nil {
			appError(w, err)
			return
		}
		switch r.Method {
		case http.MethodPost:
			h.postMethod(tx, w, r)
		case http.MethodPut:
			h.putMethod(tx, w, r, subMatches[1])
		case http.MethodDelete:
			h.deleteMethod(tx, w, r, subMatches[1])
		default:
			unknownMethod(w)
		}
	}
	m.HandleFunc(prefix, handlerFunc)
}

func appError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("error: %s", err), http.StatusInternalServerError)
}

func notImplemented(w http.ResponseWriter) {
	http.Error(w, "Not (yet) Implemented", http.StatusInternalServerError)
}

func unknownMethod(w http.ResponseWriter) {
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func serveHTTP(cfg *config.Config) {
	db, err := sqlstore.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hmux := http.NewServeMux()

	uh := newUserHandler()
	uh.register(hmux, db, "/user/")
	http.ListenAndServe(cfg.HTTPListen, hmux)
}
