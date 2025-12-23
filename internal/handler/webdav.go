package fbhttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	settings "github.com/nulnl/nulyun/internal/model/global"
	"github.com/nulnl/nulyun/internal/model/webdav"
	fberrors "github.com/nulnl/nulyun/internal/pkg_errors"
	storage "github.com/nulnl/nulyun/internal/repository"
)

// webdavTokenListHandler get all WebDAV tokens for the current user
var webdavTokenListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	tokens, err := d.store.WebDAV.GetByUserID(d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Do not return the full token string; return a partial value for identification
	for _, token := range tokens {
		if len(token.Token) > 8 {
			token.Token = token.Token[:8] + "..."
		}
	}

	return renderJSON(w, r, tokens)
})

// webdavTokenGetHandler get a single WebDAV token
var webdavTokenGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	token, err := d.store.WebDAV.Get(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	// Check whether the token belongs to the current user
	if token.UserID != d.user.ID {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	// Return the full token (only returned once when retrieved)
	return renderJSON(w, r, token)
})

// webdavTokenCreateRequest request structure for creating a token
type webdavTokenCreateRequest struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CanRead   bool   `json:"canRead"`
	CanWrite  bool   `json:"canWrite"`
	CanDelete bool   `json:"canDelete"`
}

// webdavTokenCreateHandler create a new WebDAV token
var webdavTokenCreateHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var req webdavTokenCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	if req.Name == "" {
		return http.StatusBadRequest, fberrors.ErrEmptyField
	}

	// Create a new token
	token, err := webdav.NewToken(d.user.ID, req.Name, req.Path, req.CanRead, req.CanWrite, req.CanDelete)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Save to the database
	if err := d.store.WebDAV.Save(token); err != nil {
		return http.StatusInternalServerError, err
	}

	// Return the created token (including the full token string)
	return renderJSON(w, r, token)
})

// webdavTokenUpdateRequest request structure for updating a token
type webdavTokenUpdateRequest struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CanRead   bool   `json:"canRead"`
	CanWrite  bool   `json:"canWrite"`
	CanDelete bool   `json:"canDelete"`
}

// webdavTokenUpdateHandler update a WebDAV token
var webdavTokenUpdateHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	token, err := d.store.WebDAV.Get(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	// Check whether the token belongs to the current user
	if token.UserID != d.user.ID {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	var req webdavTokenUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	// Update fields
	if req.Name != "" {
		token.Name = req.Name
	}
	token.Path = req.Path
	token.CanRead = req.CanRead
	token.CanWrite = req.CanWrite
	token.CanDelete = req.CanDelete
	token.UpdatedAt = time.Now()

	// Save the update
	if err := d.store.WebDAV.Update(token); err != nil {
		return http.StatusInternalServerError, err
	}

	// Do not return the full token
	if len(token.Token) > 8 {
		token.Token = token.Token[:8] + "..."
	}

	return renderJSON(w, r, token)
})

// webdavTokenDeleteHandler delete a WebDAV token
var webdavTokenDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	token, err := d.store.WebDAV.Get(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	// Check whether the token belongs to the current user
	if token.UserID != d.user.ID {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	if err := d.store.WebDAV.Delete(uint(id)); err != nil {
		return http.StatusInternalServerError, err
	}

	w.WriteHeader(http.StatusNoContent)
	return 0, nil
})

// webdavTokenSuspendHandler suspend a WebDAV token
var webdavTokenSuspendHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	token, err := d.store.WebDAV.Get(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	// Check whether the token belongs to the current user
	if token.UserID != d.user.ID {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	if err := d.store.WebDAV.Suspend(uint(id)); err != nil {
		return http.StatusInternalServerError, err
	}

	w.WriteHeader(http.StatusNoContent)
	return 0, nil
})

// webdavTokenActivateHandler activate a WebDAV token
var webdavTokenActivateHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	token, err := d.store.WebDAV.Get(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	// Check whether the token belongs to the current user
	if token.UserID != d.user.ID {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	if err := d.store.WebDAV.Activate(uint(id)); err != nil {
		return http.StatusInternalServerError, err
	}

	w.WriteHeader(http.StatusNoContent)
	return 0, nil
})

// setupWebDAVRoutes set up routes related to WebDAV
func setupWebDAVRoutes(api *mux.Router, store *storage.Storage, server *settings.Server) {
	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	// Token management API
	webdavAPI := api.PathPrefix("/webdav/tokens").Subrouter()
	webdavAPI.Handle("", monkey(webdavTokenListHandler, "")).Methods("GET")
	webdavAPI.Handle("", monkey(webdavTokenCreateHandler, "")).Methods("POST")
	webdavAPI.Handle("/{id:[0-9]+}", monkey(webdavTokenGetHandler, "")).Methods("GET")
	webdavAPI.Handle("/{id:[0-9]+}", monkey(webdavTokenUpdateHandler, "")).Methods("PUT")
	webdavAPI.Handle("/{id:[0-9]+}", monkey(webdavTokenDeleteHandler, "")).Methods("DELETE")
	webdavAPI.Handle("/{id:[0-9]+}/suspend", monkey(webdavTokenSuspendHandler, "")).Methods("POST")
	webdavAPI.Handle("/{id:[0-9]+}/activate", monkey(webdavTokenActivateHandler, "")).Methods("POST")
}

// setupWebDAVHandler set up the WebDAV file access handler
func setupWebDAVHandler(r *mux.Router, store *storage.Storage) {
	// WebDAV file access (no authentication middleware needed because token auth is used)
	handler := webdav.NewHandler(store.WebDAV, store.Users)
	r.PathPrefix("/dav/").Handler(handler)
}
