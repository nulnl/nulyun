package bolt

import (
	"github.com/asdine/storm/v3"

	fberrors "github.com/nulnl/nulyun/internal/pkg_errors"
	"github.com/nulnl/nulyun/internal/model/webdav"
)

type webdavBackend struct {
	db *storm.DB
}

// GetByID get token by ID
func (b webdavBackend) GetByID(id uint) (*webdav.Token, error) {
	var token webdav.Token
	err := b.db.One("ID", id, &token)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, fberrors.ErrNotExist
		}
		return nil, err
	}
	return &token, nil
}

// GetByToken get token by token string
func (b webdavBackend) GetByToken(tokenStr string) (*webdav.Token, error) {
	var token webdav.Token
	err := b.db.One("Token", tokenStr, &token)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, fberrors.ErrNotExist
		}
		return nil, err
	}
	return &token, nil
}

// GetByUserID get all tokens for a user
func (b webdavBackend) GetByUserID(userID uint) ([]*webdav.Token, error) {
	var tokens []*webdav.Token
	err := b.db.Find("UserID", userID, &tokens)
	if err != nil {
		if err == storm.ErrNotFound {
			return []*webdav.Token{}, nil
		}
		return nil, err
	}
	return tokens, nil
}

// GetAll get all tokens
func (b webdavBackend) GetAll() ([]*webdav.Token, error) {
	var tokens []*webdav.Token
	err := b.db.All(&tokens)
	if err != nil {
		if err == storm.ErrNotFound {
			return []*webdav.Token{}, nil
		}
		return nil, err
	}
	return tokens, nil
}

// Save save a new token
func (b webdavBackend) Save(token *webdav.Token) error {
	return b.db.Save(token)
}

// Update update a token
func (b webdavBackend) Update(token *webdav.Token, fields ...string) error {
	if len(fields) == 0 {
		return b.db.Update(token)
	}

	// Update multiple fields
	for _, field := range fields {
		if err := b.db.UpdateField(token, field, getField(token, field)); err != nil {
			return err
		}
	}
	return nil
}

// Delete delete a token
func (b webdavBackend) Delete(id uint) error {
	var token webdav.Token
	token.ID = id
	return b.db.DeleteStruct(&token)
}

// getField get the value of a token struct field
func getField(token *webdav.Token, field string) interface{} {
	switch field {
	case "Name":
		return token.Name
	case "Path":
		return token.Path
	case "CanRead":
		return token.CanRead
	case "CanWrite":
		return token.CanWrite
	case "CanDelete":
		return token.CanDelete
	case "Status":
		return token.Status
	case "UpdatedAt":
		return token.UpdatedAt
	default:
		return nil
	}
}
