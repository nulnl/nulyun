package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/nulnl/nulyun/auth"
	settings "github.com/nulnl/nulyun/settings/global"
	"github.com/nulnl/nulyun/settings/share"
	"github.com/nulnl/nulyun/settings/users"
	"github.com/nulnl/nulyun/settings/webdav"
	"github.com/nulnl/nulyun/storage"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)
	webdavStore := webdav.NewStorage(webdavBackend{db: db})
	passkeyStore := NewPasskeyStore(db)

	err := save(db, "version", 2)
	if err != nil {
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
		WebDAV:   webdavStore,
		Passkey:  passkeyStore,
	}, nil
}
