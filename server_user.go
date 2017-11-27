package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/model"
	"net/http"
)

func userGet(db *sql.DB, w http.ResponseWriter, r *http.Request, login string) {
	user, err := model.LoadUser(db, login)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		appError(w, err)
		return
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		appError(w, err)
		return
	}
	headers := w.Header()
	headers.Add("Content-Type", "application/json")
	fmt.Fprint(w, string(userJSON))
}

func userList(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	notImplemented(w)
}

func userCreate(db *sql.Tx, w http.ResponseWriter, r *http.Request) {
	notImplemented(w)
}

func userUpdate(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) {
	notImplemented(w)
}

func userDelete(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) {
	notImplemented(w)
}

func newUserHandler() resourceHandler {
	return resourceHandler{
		getMethod:    userGet,
		listMethod:   userList,
		putMethod:    userUpdate,
		postMethod:   userCreate,
		deleteMethod: userDelete,
	}
}
