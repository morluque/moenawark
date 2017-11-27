package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/model"
	"github.com/morluque/moenawark/sqlstore"
	"net/http"
)

func userGet(db sqlstore.DB, w http.ResponseWriter, r *http.Request, login string) {
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

func userList(db sqlstore.DB, w http.ResponseWriter, r *http.Request) {
	notImplemented(w)
}

func userCreate(db sqlstore.DB, w http.ResponseWriter, r *http.Request) {
	notImplemented(w)
}

func userUpdate(db sqlstore.DB, w http.ResponseWriter, r *http.Request, login string) {
	notImplemented(w)
}

func userDelete(db sqlstore.DB, w http.ResponseWriter, r *http.Request, login string) {
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
