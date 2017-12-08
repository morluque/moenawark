package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/model"
	"github.com/morluque/moenawark/mwkerr"
	"io"
	"net/http"
	"strconv"
)

// UserHandler is a resource handler for users.
type UserHandler struct {
	*resourceMapper
}

// SetResourceMapper sets the resourceMapper that can be used to create URLs to arbitrary resources.
func (h UserHandler) SetResourceMapper(m *resourceMapper) {
	h.resourceMapper = m
}

// View sends JSON of a user in response to HTTP GET
func (h UserHandler) View(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	user, err := getAuthUser(r)
	if err != nil {
		return authError(err)
	}
	if user.Login != login && !user.GameMaster {
		return authError(fmt.Errorf("can only get info about yourself"))
	}
	u, herr := h.loadUserFromLogin(db, login)
	if herr != nil {
		return herr
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

// List sends JSON of a list of users on HTTP GET
func (h UserHandler) List(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	user, err := getAuthUser(r)
	if err != nil {
		return authError(err)
	}
	if !user.GameMaster {
		return authError(fmt.Errorf("Only game masters can list all users"))
	}
	start, count := h.userListGetLimit(r)
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

// Create checks user-supplied JSON and creates a new user.
func (h UserHandler) Create(db *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError {
	type userCreateParams struct {
		Login     string `json:"login"`
		Password1 string `json:"password1"`
		Password2 string `json:"password2"`
	}

	data, herr := h.readBodyData(r)
	if herr != nil {
		return herr
	}

	body := userCreateParams{}
	if err := json.Unmarshal(data, &body); err != nil {
		return userError(fmt.Errorf("Error decoding JSON: %s", err.Error()))
	}
	if len(body.Login) <= 0 {
		return userError(fmt.Errorf("Login is empty"))
	}
	if body.Password1 != body.Password2 {
		return userError(fmt.Errorf("Passwords don't match"))
	}

	u := model.NewUser(body.Login, body.Password1)
	err := u.Save(db)
	if err != nil {
		merr, ok := err.(mwkerr.MWKError)
		if ok && merr.Code == mwkerr.DuplicateModel {
			return userError(err)
		}
		return appError(fmt.Errorf("Error while saving user %s: %s", body.Login, err.Error()))
	}
	err = db.Commit()
	if err != nil {
		return appError(fmt.Errorf("Database error: %s", err.Error()))
	}
	log.Infof("User %s created", u.Login)
	responseBody, err := json.Marshal(u)
	if err != nil {
		return appError(fmt.Errorf("Error encoding user %s to JSON: %s", body.Login, err.Error()))
	}
	fmt.Fprint(w, string(responseBody))

	return nil
}

// Update checks user-supplied JSON and updates a user
func (h UserHandler) Update(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	type userUpdateParams struct {
		Password1  string
		Password2  string
		Status     string
		GameMaster bool
	}

	user, err := getAuthUser(r)
	if err != nil {
		return authError(err)
	}
	if user.Login != login && !user.GameMaster {
		return authError(fmt.Errorf("can only modify yourself"))
	}

	u, herr := h.loadUserFromLogin(db, login)
	if herr != nil {
		return herr
	}

	data, herr := h.readBodyData(r)
	if herr != nil {
		return herr
	}

	nu := userUpdateParams{}
	if err = json.Unmarshal(data, &nu); err != nil {
		return userError(fmt.Errorf("Error decoding JSON: %s", err.Error()))
	}

	if user.Login == login {
		if nu.Password1 != nu.Password2 {
			return userError(fmt.Errorf("Passwords don't match"))
		}
		u.SetPassword(nu.Password1)
	} else {
		if nu.Status == "new" || nu.Status == "active" || nu.Status == "archived" {
			u.Status = nu.Status
		} else {
			return userError(fmt.Errorf("Bad user status %s, expected new, active or archived", nu.Status))
		}
		u.GameMaster = nu.GameMaster
	}

	err = u.Save(db)
	if err != nil {
		return appError(fmt.Errorf("Error saving user %s: %s", login, err.Error()))
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

// Delete would delete a user from database, but is unimplemented.
// TODO: implement for new users (users that never were active).
func (h UserHandler) Delete(db *sql.Tx, w http.ResponseWriter, r *http.Request, login string) *httpError {
	return unknownMethodError(r.Method)
}

func (h UserHandler) loadUserFromLogin(db *sql.Tx, login string) (*model.User, *httpError) {
	u, err := model.LoadUser(db, login)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, notFoundError()
		}
		return nil, appError(err)
	}
	return u, nil
}

func (h UserHandler) userListGetLimit(r *http.Request) (uint, uint) {
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

func (h UserHandler) readBodyData(r *http.Request) ([]byte, *httpError) {
	if r.ContentLength <= 0 {
		return nil, userError(fmt.Errorf("Empty request body"))
	}
	if r.ContentLength > MaxBodyLength {
		return nil, userError(fmt.Errorf("Request body too long, max length is %d", MaxBodyLength))
	}
	data := make([]byte, r.ContentLength)
	_, err := io.ReadFull(r.Body, data)
	if err != nil {
		return nil, appError(fmt.Errorf("Error while reading request body: %s", err.Error()))
	}
	return data, nil
}
