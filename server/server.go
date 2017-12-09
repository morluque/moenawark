/*
Package server implements the JSON API HTTP server of Moenawark.
*/
package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"github.com/morluque/moenawark/mwkerr"
	"github.com/morluque/moenawark/sqlstore"
	"net/http"
	"regexp"
)

const (
	// MaxBodyLength is the maximum body size in bytes that a client can send us.
	MaxBodyLength = 1024 * 1024
)

type httpError struct {
	Code    int
	Message string
	Err     error
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *httpError) MarshalJSON() ([]byte, error) {
	if err, ok := e.Err.(mwkerr.MWKError); ok {
		return json.Marshal(err)
	}
	j := fmt.Sprintf(`{"error":"%s"}`, e.Err.Error())
	return []byte(j), nil
}

type resourceHandler interface {
	List(tx *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError
	View(tx *sql.Tx, w http.ResponseWriter, r *http.Request, id string) *httpError
	Create(tx *sql.Tx, w http.ResponseWriter, r *http.Request) *httpError
	Update(tx *sql.Tx, w http.ResponseWriter, r *http.Request, id string) *httpError
	Delete(tx *sql.Tx, w http.ResponseWriter, r *http.Request, id string) *httpError
	URLTo(resourceName string, resourceID int) string
	SetResourceMapper(m *resourceMapper)
}

type resourceMapper struct {
	baseURL  string
	prefixes map[string]string
}

func newResourceMapper(baseURL string, prefixes map[string]string) *resourceMapper {
	return &resourceMapper{
		baseURL:  baseURL,
		prefixes: prefixes,
	}
}

func (m *resourceMapper) URLTo(name string, id int) string {
	prefix, ok := m.prefixes[name]
	if !ok {
		prefix = "thisisabug"
	}
	return fmt.Sprintf("%s/%s/%d", m.baseURL, prefix, id)
}

type apiServerV1 struct {
	baseURL     string
	apiPrefix   string
	apiVersion  string
	db          *sql.DB
	resourceMap map[string]string
	resources   map[string]resourceHandler
}

var log *loglevel.Logger

func init() {
	log = loglevel.New("server", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
	log.SetLevelName(config.Get("loglevel.server"))
}

func newAPIServerV1(db *sql.DB) *apiServerV1 {
	srv := apiServerV1{
		apiPrefix:   config.Get("api_prefix"),
		apiVersion:  "v1",
		db:          db,
		resourceMap: make(map[string]string),
		resources:   make(map[string]resourceHandler),
	}
	srv.baseURL = fmt.Sprintf("%s%s/%s", config.Get("base_url"), srv.apiPrefix, srv.apiVersion)
	return &srv
}

func (srv *apiServerV1) ServeMux() *http.ServeMux {
	hmux := http.NewServeMux()
	for name, handler := range srv.resources {
		prefix := fmt.Sprintf("%s/%s/%s/", srv.apiPrefix, srv.apiVersion, srv.resourceMap[name])
		hmux.HandleFunc(prefix, srv.handlerFuncFor(handler, name, prefix))
		log.Debugf("registered handlerfunc for prefix %s", prefix)
	}

	return hmux
}

func (srv *apiServerV1) handlerFuncFor(h resourceHandler, resourceName, prefix string) http.HandlerFunc {
	h.SetResourceMapper(newResourceMapper(srv.baseURL, srv.resourceMap))

	log.Debugf("making handlerfunc for prefix %s", prefix)
	reStr := fmt.Sprintf("^%s([^/]+)?$", prefix)
	re, err := regexp.Compile(reStr)
	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		// we assume an error occurred and we rollback any
		// work. We ignore any error during rollback since any
		// error would have been detected at commit time or would
		// already have occurred.
		defer tx.Rollback()

		var herr *httpError
		switch r.Method {
		case http.MethodGet:
			if len(subMatches[1]) == 0 {
				herr = h.List(tx, w, r)
			} else {
				herr = h.View(tx, w, r, subMatches[1])
			}
		case http.MethodPost:
			herr = h.Create(tx, w, r)
		case http.MethodPut:
			herr = h.Update(tx, w, r, subMatches[1])
		case http.MethodDelete:
			herr = h.Delete(tx, w, r, subMatches[1])
		default:
			herr = unknownMethodError(r.Method)
		}
		if herr != nil {
			// We are responsible to send the HTTP error to the client
			sendError(w, herr)
			return
		}
	}
}

func (srv *apiServerV1) register(resourceName, prefix string, h resourceHandler) {
	srv.resourceMap[resourceName] = prefix
	srv.resources[resourceName] = h
}

// ServeHTTP starts an HTTP server for the JSON REST API
func ServeHTTP() {
	db, err := sqlstore.Open(config.Get("db_path"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv1 := newAPIServerV1(db)
	srv1.register("user", "user", UserHandler{})
	srv1.register("auth", "auth", AuthHandler{})
	srv1.register("character", "character", CharacterHandler{})

	http.ListenAndServe(config.Get("http_listen"), srv1.ServeMux())
}

func sendError(w http.ResponseWriter, e *httpError) {
	if e.Err == nil {
		e.Err = fmt.Errorf(e.Message)
	}
	errJSON, err := json.Marshal(e)
	if err != nil {
		log.Errorf("Could not serialize error to JSON: %s", err.Error())
		log.Errorf("Original error: %s", e.Error())
		http.Error(w, "Could not serialize error to JSON", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	fmt.Fprint(w, string(errJSON))
}

func notFoundError() *httpError {
	return &httpError{Code: 404, Message: "Resource not found"}
}

func appError(err error) *httpError {
	log.Errorf(err.Error())
	return &httpError{Code: 500, Message: "Internal server error", Err: err}
}

func userError(err error) *httpError {
	log.Infof(err.Error())
	return &httpError{Code: 400, Message: "Bad request", Err: err}
}

func authError(err error) *httpError {
	log.Warnf("auth error: %s", err.Error())
	return &httpError{Code: 403, Message: "Forbidden", Err: err}
}

func unknownMethodError(method string) *httpError {
	return &httpError{Code: 405, Message: fmt.Sprintf("Method not allowed: %s", method)}
}
