package webdav

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	fberrors "github.com/nulnl/nulyun/internal/pkg_errors"
)

// TokenStatus represents the status of a token
type TokenStatus string

const (
	TokenActive    TokenStatus = "active"
	TokenSuspended TokenStatus = "suspended"
)

// Token represents a WebDAV access token
type Token struct {
	ID        uint        `storm:"id,increment" json:"id"`
	UserID    uint        `storm:"index" json:"userId"`
	Name      string      `json:"name"`
	Token     string      `storm:"unique" json:"token"`
	Path      string      `json:"path"`
	CanRead   bool        `json:"canRead"`
	CanWrite  bool        `json:"canWrite"`
	CanDelete bool        `json:"canDelete"`
	Status    TokenStatus `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func NewToken(userID uint, name, path string, canRead, canWrite, canDelete bool) (*Token, error) {
	if name == "" {
		return nil, fberrors.ErrEmptyField
	}
	if path == "" {
		path = "/"
	}

	tokenStr, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Token{
		UserID:    userID,
		Name:      name,
		Token:     tokenStr,
		Path:      path,
		CanRead:   canRead,
		CanWrite:  canWrite,
		CanDelete: canDelete,
		Status:    TokenActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (t *Token) IsActive() bool {
	return t.Status == TokenActive
}

func (t *Token) Suspend() {
	t.Status = TokenSuspended
	t.UpdatedAt = time.Now()
}

func (t *Token) Activate() {
	t.Status = TokenActive
	t.UpdatedAt = time.Now()
}

func (t *Token) HasPermission(read, write, del bool) bool {
	if !t.IsActive() {
		return false
	}
	if read && !t.CanRead {
		return false
	}
	if write && !t.CanWrite {
		return false
	}
	if del && !t.CanDelete {
		return false
	}
	return true
}
