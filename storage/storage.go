package storage

import (
	"github.com/nulnl/nulyun/auth"
	settings "github.com/nulnl/nulyun/settings/global"
	"github.com/nulnl/nulyun/settings/passkey"
	"github.com/nulnl/nulyun/settings/share"
	"github.com/nulnl/nulyun/settings/users"
	"github.com/nulnl/nulyun/settings/webdav"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    users.Store
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
	WebDAV   *webdav.Storage
	Passkey  passkey.Store
}
