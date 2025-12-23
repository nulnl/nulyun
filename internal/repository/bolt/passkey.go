package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/nulnl/nulyun/internal/model/passkey"
)

// PasskeyStore is the bolt implementation of passkey.Store
type PasskeyStore struct {
	db *storm.DB
}

// NewPasskeyStore creates a new PasskeyStore
func NewPasskeyStore(db *storm.DB) *PasskeyStore {
	return &PasskeyStore{db: db}
}

// Save saves a credential
func (s *PasskeyStore) Save(cred *passkey.Credential) error {
	return s.db.Save(cred)
}

// Get gets a credential by ID
func (s *PasskeyStore) Get(id uint) (*passkey.Credential, error) {
	var cred passkey.Credential
	err := s.db.One("ID", id, &cred)
	return &cred, err
}

// GetByUserID gets all credentials for a user
func (s *PasskeyStore) GetByUserID(userID uint) ([]*passkey.Credential, error) {
	var creds []*passkey.Credential
	err := s.db.Find("UserID", userID, &creds)
	if err == storm.ErrNotFound {
		return []*passkey.Credential{}, nil
	}
	return creds, err
}

// GetByCredentialID gets a credential by its credential ID
func (s *PasskeyStore) GetByCredentialID(credentialID []byte) (*passkey.Credential, error) {
	var creds []*passkey.Credential
	err := s.db.All(&creds)
	if err != nil {
		return nil, err
	}

	for _, cred := range creds {
		if string(cred.CredentialID) == string(credentialID) {
			return cred, nil
		}
	}

	return nil, storm.ErrNotFound
}

// Update updates a credential
func (s *PasskeyStore) Update(cred *passkey.Credential, fields ...string) error {
	return s.db.Update(cred)
}

// Delete deletes a credential
func (s *PasskeyStore) Delete(id uint) error {
	return s.db.DeleteStruct(&passkey.Credential{ID: id})
}
