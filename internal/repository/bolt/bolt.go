package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/nulnl/nulyun/internal/auth"
	settings "github.com/nulnl/nulyun/internal/model/global"
	"github.com/nulnl/nulyun/internal/model/share"
	"github.com/nulnl/nulyun/internal/model/users"
	"github.com/nulnl/nulyun/internal/model/webdav"
	storage "github.com/nulnl/nulyun/internal/repository"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)
	webdavStore := webdav.NewStorage(webdavBackend{db: db})

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
	}, nil
}
