package main

import (
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/sqlstore"
	"log"
	"net/http"
	"regexp"
)

type resourceMethod1 func(sqlstore.DB, http.ResponseWriter, *http.Request, string)
type resourceMethod0 func(sqlstore.DB, http.ResponseWriter, *http.Request)

type resourceHandler struct {
	listMethod   resourceMethod0
	postMethod   resourceMethod0
	getMethod    resourceMethod1
	putMethod    resourceMethod1
	deleteMethod resourceMethod1
}

func (h resourceHandler) register(m *http.ServeMux, db sqlstore.DB, prefix string) {
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
		switch r.Method {
		case http.MethodGet:
			if len(subMatches[1]) == 0 {
				h.listMethod(db, w, r)
			} else {
				h.getMethod(db, w, r, subMatches[1])
			}
		case http.MethodPost:
			h.postMethod(db, w, r)
		case http.MethodPut:
			h.putMethod(db, w, r, subMatches[1])
		case http.MethodDelete:
			h.deleteMethod(db, w, r, subMatches[1])
		default:
			unknownMethod(w)
			return
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
