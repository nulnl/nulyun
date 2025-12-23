package webdav

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"

	"github.com/nulnl/nulyun/settings/users"
)

// Handler is the WebDAV handler
type Handler struct {
	storage *Storage
	users   users.Store
}

// NewHandler creates a new WebDAV handler
func NewHandler(storage *Storage, userStore users.Store) *Handler {
	return &Handler{
		storage: storage,
		users:   userStore,
	}
}

// ServeHTTP handles WebDAV requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, tokenStr := extractToken(r)
	if username == "" || tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.users.Get("", username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	token, err := h.storage.GetByToken(tokenStr)
	if err != nil || token.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !token.IsActive() {
		http.Error(w, "Token is suspended", http.StatusForbidden)
		return
	}
	rootPath := filepath.Join(user.Scope, token.Path)
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}
	method := r.Method
	if !h.checkPermission(token, method) {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}
	handler := &webdav.Handler{
		Prefix:     "/dav",
		FileSystem: webdav.Dir(rootPath),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			// Log errors if needed
			_ = err
		},
	}
	handler.ServeHTTP(w, r)
}

func extractToken(r *http.Request) (string, string) {
	if username, password, ok := r.BasicAuth(); ok {
		return username, password
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		parts := strings.SplitN(auth, "|", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return r.URL.Query().Get("username"), r.URL.Query().Get("token")
}

func (h *Handler) checkPermission(token *Token, method string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS", "PROPFIND":
		return token.CanRead
	case "PUT", "POST", "PATCH", "MKCOL", "COPY", "MOVE":
		return token.CanWrite
	case "DELETE":
		return token.CanDelete
	default:
		return false
	}
}

type FileInfo struct {
	os.FileInfo
}

type contextKey string

const tokenContextKey contextKey = "webdav-token"

func WithToken(ctx context.Context, token *Token) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

func GetToken(ctx context.Context) (*Token, bool) {
	token, ok := ctx.Value(tokenContextKey).(*Token)
	return token, ok
}
