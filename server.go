package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/model"
	"github.com/morluque/moenawark/sqlstore"
	"log"
	"net/http"
	"strings"
)

type userHandler struct {
	db *sql.DB
}

func sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "error: %s", err)
}

func (h userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 || len(path[2]) == 0 {
		http.NotFound(w, r)
		return
	}
	login := path[2]
	user, err := model.LoadUser(h.db, login)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		sendError(w, err)
		return
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		sendError(w, err)
		return
	}
	headers := w.Header()
	headers.Add("Content-Type", "application/json")
	fmt.Fprint(w, string(userJSON))
}

func serveHTTP(cfg *config.Config) {
	db, err := sqlstore.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	uh := userHandler{db: db}
	http.Handle("/user/", uh)
	http.ListenAndServe(cfg.HTTPListen, nil)
}
