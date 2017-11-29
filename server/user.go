package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/model"
	"net/http"
)

func userGet(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	user, err := getAuthUser(r)
	if err != nil {
		return authError(err)
	}
	if user.Login != login && !user.GameMaster {
		return authError(fmt.Errorf("can only get info about yourself"))
	}
	u, err := model.LoadUser(db, login)
	if err != nil {
		if err == sql.ErrNoRows {
			return notFoundError()
		}
		return appError(err)
	}
	userJSON, err := json.Marshal(u)
	if err != nil {
		return appError(err)
	}
	headers := w.Header()
	headers.Add("Content-Type", "application/json")
	fmt.Fprint(w, string(userJSON))

	return nil
}

func userList(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	return unknownMethodError(r.Method)
}

func userCreate(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	return unknownMethodError(r.Method)
}

func userUpdate(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
}

func userDelete(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
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
