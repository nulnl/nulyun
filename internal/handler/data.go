package fbhttp

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tomasen/realip"

	settings "github.com/nulnl/nulyun/internal/model/global"
	"github.com/nulnl/nulyun/internal/model/users"
	storage "github.com/nulnl/nulyun/internal/repository"
)

type handleFunc func(w http.ResponseWriter, r *http.Request, d *data) (int, error)

type data struct {
	settings *settings.Settings
	server   *settings.Server
	store    *storage.Storage
	user     *users.User
	raw      interface{}
}

// Check implements files.Checker.
func (d *data) Check(path string) bool {
	// If user is not set, allow access (should not happen in normal flow)
	if d.user == nil {
		return true
	}

	// Get the base name of the path
	name := filepath.Base(path)

	// Check if it's a hidden file (starts with .)
	isHidden := strings.HasPrefix(name, ".")

	if !isHidden {
		return true
	}

	// If it's hidden, check file system to determine if it's a file or directory
	if d.user.Fs != nil {
		info, err := d.user.Fs.Stat(path)
		if err == nil {
			// If it's a directory and hideHiddenFolders is enabled
			if info.IsDir() && d.user.HideHiddenFolders {
				return false
			}
			// If it's a file and hideDotfiles is enabled
			if !info.IsDir() && d.user.HideDotfiles {
				return false
			}
		}
	}

	// Default behavior: check hideDotfiles for all hidden items
	// if we couldn't determine the type
	if d.user.HideDotfiles {
		return false
	}

	return true
}

func handle(fn handleFunc, prefix string, store *storage.Storage, server *settings.Server) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range globalHeaders {
			w.Header().Set(k, v)
		}

		settings, err := store.Settings.Get()
		if err != nil {
			log.Fatalf("ERROR: couldn't get settings: %v\n", err)
			return
		}

		status, err := fn(w, r, &data{
			store:    store,
			settings: settings,
			server:   server,
		})

		if status >= 400 || err != nil {
			clientIP := realip.FromRequest(r)
			log.Printf("%s: %v %s %v", r.URL.Path, status, clientIP, err)
		}

		if status != 0 {
			txt := http.StatusText(status)
			if status == http.StatusBadRequest && err != nil {
				txt += " (" + err.Error() + ")"
			}
			http.Error(w, strconv.Itoa(status)+" "+txt, status)
			return
		}
	})

	return stripPrefix(prefix, handler)
}
