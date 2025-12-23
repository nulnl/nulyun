package fbhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"

	fberrors "github.com/nulnl/nulyun/internal/errors"
	"github.com/nulnl/nulyun/internal/settings/passkey"
	"github.com/nulnl/nulyun/internal/settings/users"
)

// WebAuthnUser implements the webauthn.User interface
type WebAuthnUser struct {
	user        *users.User
	credentials []*passkey.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(fmt.Sprintf("%d", u.user.ID))
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.user.Username
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.user.Username
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.credentials))
	for i, c := range u.credentials {
		creds[i] = c.ToWebAuthnCredential()
	}
	return creds
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// passkeyListHandler lists all passkeys for the current user
var passkeyListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Check if passkey is enabled at settings and user level
	if !d.settings.PasskeyEnabled || !d.user.PasskeyEnabled {
		return http.StatusForbidden, fmt.Errorf("passkey is not enabled")
	}

	credentials, err := d.store.Passkey.GetByUserID(d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Don't return sensitive data
	type credentialResponse struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		CreatedAt  string `json:"createdAt"`
		LastUsedAt string `json:"lastUsedAt"`
	}

	response := make([]credentialResponse, len(credentials))
	for i, cred := range credentials {
		response[i] = credentialResponse{
			ID:         cred.ID,
			Name:       cred.Name,
			CreatedAt:  cred.CreatedAt.Format("2006-01-02 15:04:05"),
			LastUsedAt: cred.LastUsedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return renderJSON(w, r, response)
})

// passkeyRegisterBeginHandler begins passkey registration
var passkeyRegisterBeginHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Check if passkey is enabled at settings and user level
	if !d.settings.PasskeyEnabled || !d.user.PasskeyEnabled {
		return http.StatusForbidden, fmt.Errorf("passkey is not enabled")
	}

	credentials, err := d.store.Passkey.GetByUserID(d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	webAuthnUser := &WebAuthnUser{
		user:        d.user,
		credentials: credentials,
	}

	options, sessionData, err := d.webAuthn.BeginRegistration(webAuthnUser)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// BeginRegistration options returned to client

	// Store session data in the session store (using cookie or similar)
	// For simplicity, we'll store it in a map keyed by user ID
	// In production, use a proper session store
	d.passkeyRegistrationSessions[d.user.ID] = sessionData

	return renderJSON(w, r, options)
})

// passkeyRegisterFinishHandler finishes passkey registration
var passkeyRegisterFinishHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.server.EnablePasskey {
		return http.StatusForbidden, fmt.Errorf("passkey feature is disabled")
	}

	sessionData, ok := d.passkeyRegistrationSessions[d.user.ID]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("no registration session found")
	}

	credentials, err := d.store.Passkey.GetByUserID(d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	webAuthnUser := &WebAuthnUser{
		user:        d.user,
		credentials: credentials,
	}

	// Read body first so we can parse name and also pass the body to FinishRegistration
	bodyBytes, _ := io.ReadAll(r.Body)
	// restore body for FinishRegistration
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	credential, err := d.webAuthn.FinishRegistration(webAuthnUser, *sessionData, r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Get credential name from request (if provided)
	type registerRequest struct {
		Name string `json:"name"`
	}
	var req registerRequest
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			req.Name = "Passkey"
		}
	}
	if req.Name == "" {
		req.Name = "Passkey"
	}

	// Save credential
	cred := passkey.FromWebAuthnCredential(d.user.ID, req.Name, *credential)
	if err := d.store.Passkey.Save(cred); err != nil {
		return http.StatusInternalServerError, err
	}

	// Clean up session
	delete(d.passkeyRegistrationSessions, d.user.ID)

	return renderJSON(w, r, cred)
})

// passkeyDeleteHandler deletes a passkey
var passkeyDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.server.EnablePasskey {
		return http.StatusForbidden, fmt.Errorf("passkey feature is disabled")
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	credential, err := d.store.Passkey.Get(uint(id))
	if err != nil {
		return http.StatusNotFound, err
	}

	if credential.UserID != d.user.ID {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	if err := d.store.Passkey.Delete(uint(id)); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})

// passkeyLoginBeginHandler begins passkey login
var passkeyLoginBeginHandler = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.server.EnablePasskey {
		return http.StatusForbidden, fmt.Errorf("passkey feature is disabled")
	}

	options, sessionData, err := d.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Store session data
	// For a public endpoint, we need a different approach
	// We'll use a session cookie or similar
	// For now, store in memory with a generated session ID
	sessionID := string(sessionData.Challenge)
	d.passkeyLoginSessions[sessionID] = sessionData

	// Set session ID in response header
	w.Header().Set("X-Passkey-Session-ID", sessionID)

	return renderJSON(w, r, options)
}

// passkeyLoginFinishHandler finishes passkey login
var passkeyLoginFinishHandler = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.server.EnablePasskey {
		return http.StatusForbidden, fmt.Errorf("passkey feature is disabled")
	}

	sessionID := r.Header.Get("X-Passkey-Session-ID")
	if sessionID == "" {
		return http.StatusBadRequest, fmt.Errorf("no session ID provided")
	}

	sessionData, ok := d.passkeyLoginSessions[sessionID]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("no login session found")
	}

	// Parse the credential from request
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Find the credential
	credential, err := d.store.Passkey.GetByCredentialID(parsedResponse.RawID)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("credential not found")
	}

	// Get the user
	user, err := d.store.Users.Get(d.server.Root, credential.UserID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Check if passkey login is enabled (server, settings, and user level)
	if !d.server.EnablePasskey || !d.settings.PasskeyEnabled || !user.PasskeyEnabled {
		return http.StatusForbidden, fmt.Errorf("passkey login is not enabled")
	}

	// Get all user credentials for validation
	credentials, err := d.store.Passkey.GetByUserID(user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	webAuthnUser := &WebAuthnUser{
		user:        user,
		credentials: credentials,
	}

	// Finish login
	_, err = d.webAuthn.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
		return webAuthnUser, nil
	}, *sessionData, parsedResponse)

	if err != nil {
		return http.StatusUnauthorized, err
	}

	// Update last used time
	credential.LastUsedAt = time.Now()
	if err := d.store.Passkey.Update(credential, "LastUsedAt"); err != nil {
		return http.StatusInternalServerError, err
	}

	// Clean up session
	delete(d.passkeyLoginSessions, sessionID)

	// Generate auth token
	tokenExpireTime, _ := d.server.GetTokenExpirationTime(DefaultTokenExpirationTime, DefaultTOTPTokenExpirationTime)
	return printToken(w, r, d, user, tokenExpireTime)
}
