package webdav

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"

	settings "github.com/nulnl/nulyun/internal/model/global"
	"github.com/nulnl/nulyun/internal/model/users"
)

// Handler is the WebDAV handler
type Handler struct {
	storage *Storage
	users   users.Store
	baseURL string
	server  *settings.Server
}

// NewHandler creates a new WebDAV handler
func NewHandler(storage *Storage, userStore users.Store, server *settings.Server) *Handler {
	return &Handler{
		storage: storage,
		users:   userStore,
		baseURL: strings.TrimSuffix(server.BaseURL, "/"),
		server:  server,
	}
}

// ServeHTTP handles WebDAV requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, tokenStr := extractToken(r)
	if username == "" || tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.users.Get(h.server.Root, username)
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

	// Calculate the actual filesystem path
	// user.Scope is the relative path from the database
	// We need to join it with server.Root to get the full path
	userFullScope := filepath.Join(h.server.Root, filepath.Join("/", user.Scope))

	// Apply token's path restriction if specified
	// token.Path is relative to the user's scope
	tokenPath := filepath.Join("/", token.Path)
	webdavRoot := filepath.Join(userFullScope, tokenPath)

	// Verify the path exists
	if _, err := os.Stat(webdavRoot); os.IsNotExist(err) {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	// Create WebDAV handler with golang.org/x/net/webdav
	// Prefix must include BaseURL so responses contain correct paths
	// Use the token's restricted path
	mountPath := h.baseURL + "/dav"
	handler := &webdav.Handler{
		Prefix:     mountPath,
		FileSystem: webdav.Dir(webdavRoot),
		LockSystem: webdav.NewMemLS(),
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
