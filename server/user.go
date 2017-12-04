package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/model"
	"io"
	"net/http"
	"strconv"
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

func userListGetLimit(r *http.Request) (uint, uint) {
	var (
		start uint
		count uint = 100
	)
	startStr := r.FormValue("start")
	if len(startStr) <= 0 {
		i, err := strconv.Atoi(startStr)
		if err == nil && i >= 0 {
			start = uint(i)
		}
	}
	countStr := r.FormValue("count")
	if len(countStr) <= 0 {
		i, err := strconv.Atoi(countStr)
		if err == nil && i > 0 {
			count = uint(i)
		}
	}
	return start, count
}

func userList(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	user, err := getAuthUser(r)
	if err != nil {
		return authError(err)
	}
	if !user.GameMaster {
		return authError(fmt.Errorf("Only game masters can list all users"))
	}
	start, count := userListGetLimit(r)
	users, err := model.ListUsers(db, start, count)
	if err != nil {
		return appError(err)
	}
	userJSON, err := json.Marshal(users)
	if err != nil {
		return appError(err)
	}
	headers := w.Header()
	headers.Add("Content-Type", "application/json")
	fmt.Fprint(w, string(userJSON))

	return nil
}

type userCreateParams struct {
	Login     string `json:"login"`
	Password1 string `json:"password1"`
	Password2 string `json:"password2"`
}

func userCreate(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	if r.ContentLength <= 0 {
		return userError(fmt.Errorf("Empty request body"))
	}
	if r.ContentLength > MaxBodyLength {
		return userError(fmt.Errorf("Request body too long, max length is %d", MaxBodyLength))
	}
	data := make([]byte, r.ContentLength)
	_, err := io.ReadFull(r.Body, data)
	if err != nil {
		return appError(fmt.Errorf("Error while reading request body: %s", err.Error()))
	}
	body := userCreateParams{}
	if err = json.Unmarshal(data, body); err != nil {
		return userError(fmt.Errorf("Error decoding JSON: %s", err.Error()))
	}
	if len(body.Login) <= 0 {
		return userError(fmt.Errorf("Login is empty"))
	}
	if body.Password1 != body.Password2 {
		return userError(fmt.Errorf("Passwords don't match"))
	}
	u := model.NewUser(body.Login, body.Password1)
	err = u.Save(db)
	if err != nil {
		return appError(fmt.Errorf("Error while saving user %s: %s", body.Login, err.Error()))
	}
	err = db.Commit()
	if err != nil {
		return appError(fmt.Errorf("Database error: %s", err.Error()))
	}
	responseBody, err := json.Marshal(u)
	if err != nil {
		return appError(fmt.Errorf("Error encoding user %s to JSON: %s", body.Login, err.Error()))
	}
	fmt.Fprint(w, string(responseBody))

	return nil
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
