package storage

import (
	"github.com/nulnl/nulyun/internal/auth"
	settings "github.com/nulnl/nulyun/internal/model/global"
	"github.com/nulnl/nulyun/internal/model/share"
	"github.com/nulnl/nulyun/internal/model/users"
	"github.com/nulnl/nulyun/internal/model/webdav"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    users.Store
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
	WebDAV   *webdav.Storage
}
