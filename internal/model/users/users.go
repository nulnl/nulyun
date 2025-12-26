package users

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/nulnl/nulyun/internal/files"
	fberrors "github.com/nulnl/nulyun/internal/pkg_errors"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// User describes a user.
type User struct {
	ID             uint          `storm:"id,increment" json:"id"`
	Username       string        `storm:"unique" json:"username"`
	Password       string        `json:"password"`
	Scope          string        `json:"scope"`
	Locale         string        `json:"locale"`
	LockPassword   bool          `json:"lockPassword"`
	ViewMode       ViewMode      `json:"viewMode"`
	SingleClick    bool          `json:"singleClick"`
	Perm           Permissions   `json:"perm"`
	Sorting        files.Sorting `json:"sorting"`
	Fs             afero.Fs      `json:"-" yaml:"-"`
	HideDotfiles   bool          `json:"hideDotfiles"`
	DateFormat     bool          `json:"dateFormat"`
	AceEditorTheme string        `json:"aceEditorTheme"`
	TOTPSecret     string        `json:"totpSecret"`
	TOTPNonce      string        `json:"totpNonce"`
	TOTPVerified   bool          `json:"totpVerified"`
	TOTPEnabled    bool          `json:"totpEnabled"`
	RecoveryCodes  []string      `json:"recoveryCodes"`
	StorageQuota   int64         `json:"storageQuota"` // in bytes, 0 means unlimited
}

var checkableFields = []string{
	"Username",
	"Password",
	"Scope",
	"ViewMode",
	"Sorting",
}

// Clean cleans up a user and verifies if all its fields
// are alright to be saved.
func (u *User) Clean(baseScope string, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return fberrors.ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return fberrors.ErrEmptyPassword
			}
		case "ViewMode":
			if u.ViewMode == "" {
				u.ViewMode = ListViewMode
			}
		case "Sorting":
			if u.Sorting.By == "" {
				u.Sorting.By = "name"
			}
		}
	}

	if u.Fs == nil {
		scope := u.Scope
		scope = filepath.Join(baseScope, filepath.Join("/", scope))
		u.Fs = afero.NewBasePathFs(afero.NewOsFs(), scope)
	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {
	return afero.FullBaseFsPath(u.Fs.(*afero.BasePathFs), path)
}
