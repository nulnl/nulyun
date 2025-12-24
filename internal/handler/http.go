package fbhttp

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"

	settings "github.com/nulnl/nulyun/internal/model/global"
	"github.com/nulnl/nulyun/internal/model/webdav"
	storage "github.com/nulnl/nulyun/internal/repository"
)

type modifyRequest struct {
	What  string   `json:"what"`  // Answer to: what data type?
	Which []string `json:"which"` // Answer to: which fields?
}

var (
	globalWebAuthn                    *webauthn.WebAuthn
	globalPasskeyRegistrationSessions = make(map[uint]*webauthn.SessionData)
	globalPasskeyLoginSessions        = make(map[string]*webauthn.SessionData)
)

func NewHandler(
	imgSvc ImgService,
	fileCache FileCache,
	store *storage.Storage,
	server *settings.Server,
	assetsFs fs.FS,
) (http.Handler, error) {
	server.Clean()

	// Initialize WebAuthn
	// The RP ID should be the domain without protocol/port
	// For simplicity, we'll use "localhost" for dev and rely on the frontend to provide correct origin
	wconfig := &webauthn.Config{
		RPDisplayName: "Nul Yun",
		RPID:          "localhost",                                            // Will be overridden by request origin
		RPOrigins:     []string{"http://localhost:8080", "https://localhost"}, // Development defaults
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}
	globalWebAuthn = webAuthn

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
			next.ServeHTTP(w, r)
		})
	})
	index, static := getStaticHandlers(store, server, assetsFs)

	// NOTE: This fixes the issue where it would redirect if people did not put a
	// trailing slash in the end. I hate this decision since this allows some awful
	// URLs https://www.gorillatoolkit.org/pkg/mux#Router.SkipClean
	r = r.SkipClean(true)

	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	r.HandleFunc("/health", healthHandler)
	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	tokenExpirationTime, totpExpTime := server.GetTokenExpirationTime(DefaultTokenExpirationTime, DefaultTOTPTokenExpirationTime)
	api.Handle("/login", monkey(loginHandler(tokenExpirationTime, totpExpTime), ""))
	api.Handle("/login/otp", monkey(verifyTOTPHandler(tokenExpirationTime), ""))
	api.Handle("/signup", monkey(signupHandler, ""))
	api.Handle("/renew", monkey(renewHandler(tokenExpirationTime), ""))

	users := api.PathPrefix("/users").Subrouter()
	users.Handle("", monkey(usersGetHandler, "")).Methods("GET")
	users.Handle("", monkey(userPostHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}", monkey(userPutHandler, "")).Methods("PUT")
	users.Handle("/{id:[0-9]+}", monkey(userGetHandler, "")).Methods("GET")
	users.Handle("/{id:[0-9]+}", monkey(userDeleteHandler, "")).Methods("DELETE")
	users.Handle("/{id:[0-9]+}/otp", monkey(userEnableTOTPHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}/otp", monkey(userGetTOTPHandler, "")).Methods("GET")
	users.Handle("/{id:[0-9]+}/otp/check", monkey(userCheckTOTPHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}/otp", monkey(userDisableTOTPHandler, "")).Methods("DELETE")
	users.Handle("/{id:[0-9]+}/otp/reset", monkey(userResetTOTPHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}/otp/recovery", monkey(userGenerateRecoveryCodesHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}/otp/toggle", monkey(userToggleTOTPHandler, "")).Methods("PUT")
	users.Handle("/{id:[0-9]+}/passkey/toggle", monkey(userTogglePasskeyHandler, "")).Methods("PUT")

	// Passkey routes
	passkeys := api.PathPrefix("/passkeys").Subrouter()
	passkeys.Handle("", monkey(passkeyListHandler, "")).Methods("GET")
	passkeys.Handle("/register/begin", monkey(passkeyRegisterBeginHandler, "")).Methods("POST")
	passkeys.Handle("/register/finish", monkey(passkeyRegisterFinishHandler, "")).Methods("POST")
	passkeys.Handle("/{id:[0-9]+}", monkey(passkeyDeleteHandler, "")).Methods("DELETE")

	// Public passkey login endpoints
	api.Handle("/passkey/login/begin", monkey(handleFunc(passkeyLoginBeginHandler), "")).Methods("POST")
	api.Handle("/passkey/login/finish", monkey(handleFunc(passkeyLoginFinishHandler), "")).Methods("POST")

	api.PathPrefix("/resources").Handler(monkey(resourceGetHandler, "/api/resources")).Methods("GET")
	api.PathPrefix("/resources").Handler(monkey(resourceDeleteHandler(fileCache), "/api/resources")).Methods("DELETE")
	api.PathPrefix("/resources").Handler(monkey(resourcePostHandler(fileCache), "/api/resources")).Methods("POST")
	api.PathPrefix("/resources").Handler(monkey(resourcePutHandler, "/api/resources")).Methods("PUT")
	api.PathPrefix("/resources").Handler(monkey(resourcePatchHandler(fileCache), "/api/resources")).Methods("PATCH")

	api.PathPrefix("/tus").Handler(monkey(tusPostHandler(), "/api/tus")).Methods("POST")
	api.PathPrefix("/tus").Handler(monkey(tusHeadHandler(), "/api/tus")).Methods("HEAD", "GET")
	api.PathPrefix("/tus").Handler(monkey(tusPatchHandler(), "/api/tus")).Methods("PATCH")
	api.PathPrefix("/tus").Handler(monkey(tusDeleteHandler(), "/api/tus")).Methods("DELETE")

	api.PathPrefix("/usage").Handler(monkey(diskUsage, "/api/usage")).Methods("GET")

	api.Path("/shares").Handler(monkey(shareListHandler, "/api/shares")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(shareGetsHandler, "/api/share")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(sharePostHandler, "/api/share")).Methods("POST")
	api.PathPrefix("/share").Handler(monkey(shareDeleteHandler, "/api/share")).Methods("DELETE")

	api.Handle("/settings", monkey(settingsGetHandler, "")).Methods("GET")
	api.Handle("/settings", monkey(settingsPutHandler, "")).Methods("PUT")

	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	api.PathPrefix("/preview/{size}/{path:.*}").
		Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")
	api.PathPrefix("/subtitle").Handler(monkey(subtitleHandler, "/api/subtitle")).Methods("GET")

	public := api.PathPrefix("/public").Subrouter()
	public.PathPrefix("/dl").Handler(monkey(publicDlHandler, "/api/public/dl/")).Methods("GET")
	public.PathPrefix("/share").Handler(monkey(publicShareHandler, "/api/public/share/")).Methods("GET")

	// WebDAV routes
	setupWebDAVRoutes(api, store, server)
	
	// Create a wrapper handler that processes WebDAV separately
	// WebDAV needs to see the full path including BaseURL for correct response generation
	wrapper := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check if this is a WebDAV request (including BaseURL)
		davPath := server.BaseURL + "/dav/"
		if strings.HasPrefix(req.URL.Path, davPath) {
			// Handle WebDAV directly without stripPrefix
			webdavHandler := webdav.NewHandler(store.WebDAV, store.Users, server.BaseURL)
			webdavHandler.ServeHTTP(w, req)
			return
		}
		// For all other routes, apply stripPrefix
		stripPrefix(server.BaseURL, r).ServeHTTP(w, req)
	})

	return wrapper, nil
}
