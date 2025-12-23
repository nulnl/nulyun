package webdav

import (
	"sync"

	fberrors "github.com/nulnl/nulyun/internal/errors"
)

// StorageBackend is the interface for WebDAV token storage
type StorageBackend interface {
	GetByID(id uint) (*Token, error)
	GetByToken(token string) (*Token, error)
	GetByUserID(userID uint) ([]*Token, error)
	GetAll() ([]*Token, error)
	Save(token *Token) error
	Update(token *Token, fields ...string) error
	Delete(id uint) error
}

// Storage is the storage manager for WebDAV tokens
type Storage struct {
	back StorageBackend
	mux  sync.RWMutex
}

// NewStorage creates a new WebDAV token storage manager
func NewStorage(back StorageBackend) *Storage {
	return &Storage{
		back: back,
	}
}

// Get get token by ID
func (s *Storage) Get(id uint) (*Token, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.back.GetByID(id)
}

// GetByToken get token by token string
func (s *Storage) GetByToken(token string) (*Token, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.back.GetByToken(token)
}

// GetByUserID get all tokens for a user
func (s *Storage) GetByUserID(userID uint) ([]*Token, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.back.GetByUserID(userID)
}

// GetAll get all tokens
func (s *Storage) GetAll() ([]*Token, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.back.GetAll()
}

// Save save a new token
func (s *Storage) Save(token *Token) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if token.Name == "" {
		return fberrors.ErrEmptyField
	}

	return s.back.Save(token)
}

// Update update a token
func (s *Storage) Update(token *Token, fields ...string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if token.Name == "" {
		return fberrors.ErrEmptyField
	}

	return s.back.Update(token, fields...)
}

// Delete delete a token
func (s *Storage) Delete(id uint) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.back.Delete(id)
}

// Suspend suspend a token
func (s *Storage) Suspend(id uint) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	token, err := s.back.GetByID(id)
	if err != nil {
		return err
	}

	token.Suspend()
	return s.back.Update(token, "Status", "UpdatedAt")
}

// Activate activate a token
func (s *Storage) Activate(id uint) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	token, err := s.back.GetByID(id)
	if err != nil {
		return err
	}

	token.Activate()
	return s.back.Update(token, "Status", "UpdatedAt")
}
