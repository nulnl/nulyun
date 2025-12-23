package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/spf13/afero"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/nulnl/nulyun/internal/auth"
	"github.com/nulnl/nulyun/internal/files"
	fbhttp "github.com/nulnl/nulyun/internal/http"
	settings "github.com/nulnl/nulyun/internal/settings/global"
	"github.com/nulnl/nulyun/internal/settings/users"
	"github.com/nulnl/nulyun/internal/storage"
	"github.com/nulnl/nulyun/internal/storage/bolt"
	"github.com/nulnl/nulyun/www"
)

var (
	// Config file
	configFile = flag.String("config", "", "path to JSON configuration file")

	// Database
	database = flag.String("database", "./nulyun.db", "database path")

	// Server flags
	address  = flag.String("address", "127.0.0.1", "address to listen on")
	port     = flag.String("port", "8080", "port to listen on")
	cert     = flag.String("cert", "", "tls certificate")
	key      = flag.String("key", "", "tls key")
	root     = flag.String("root", ".", "root to prepend to relative paths")
	baseURL  = flag.String("baseURL", "", "base url")
	logPath  = flag.String("log", "stdout", "log output")
	cacheDir = flag.String("cacheDir", "", "file cache directory (disabled if empty)")

	// Token settings
	tokenExpirationTime     = flag.String("tokenExpirationTime", "2h", "user session timeout")
	totpTokenExpirationTime = flag.String("totpTokenExpirationTime", "2m", "user totp session timeout to login")

	// Feature flags
	disableThumbnails            = flag.Bool("disableThumbnails", false, "disable image thumbnails")
	disablePreviewResize         = flag.Bool("disablePreviewResize", false, "disable resize of image previews")
	disableTypeDetectionByHeader = flag.Bool("disableTypeDetectionByHeader", false, "disables type detection by reading file headers")
	disableTOTP                  = flag.Bool("disableTOTP", false, "disable TOTP authentication feature")
	disablePasskey               = flag.Bool("disablePasskey", false, "disable Passkey/WebAuthn authentication feature")

	// Quick setup flags
	noauth   = flag.Bool("noauth", false, "use the noauth auther when using quick setup")
	username = flag.String("username", "admin", "username for the first user when using quick setup")
	password = flag.String("password", "", "hashed password for the first user when using quick setup")

	// Other
	imageProcessors = flag.Int("imageProcessors", 4, "image processors count")
)

// Config represents the JSON configuration file structure
type Config struct {
	Database                     string `json:"database,omitempty"`
	Address                      string `json:"address,omitempty"`
	Port                         string `json:"port,omitempty"`
	Cert                         string `json:"cert,omitempty"`
	Key                          string `json:"key,omitempty"`
	Root                         string `json:"root,omitempty"`
	BaseURL                      string `json:"baseURL,omitempty"`
	Log                          string `json:"log,omitempty"`
	CacheDir                     string `json:"cacheDir,omitempty"`
	TokenExpirationTime          string `json:"tokenExpirationTime,omitempty"`
	TotpTokenExpirationTime      string `json:"totpTokenExpirationTime,omitempty"`
	DisableThumbnails            *bool  `json:"disableThumbnails,omitempty"`
	DisablePreviewResize         *bool  `json:"disablePreviewResize,omitempty"`
	DisableTypeDetectionByHeader *bool  `json:"disableTypeDetectionByHeader,omitempty"`
	DisableTOTP                  *bool  `json:"disableTOTP,omitempty"`
	DisablePasskey               *bool  `json:"disablePasskey,omitempty"`
	Noauth                       *bool  `json:"noauth,omitempty"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	ImageProcessors              *int   `json:"imageProcessors,omitempty"`
}

func main() {
	flag.Parse()

	// Load config file if specified
	if *configFile != "" {
		if err := loadConfig(*configFile); err != nil {
			log.Fatal(err)
		}
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Check if database exists
	databaseExisted, err := dbExists(*database)
	if err != nil {
		return err
	}

	// Open database
	db, err := storm.Open(*database, storm.BoltOptions(0640, nil))
	if err != nil {
		return err
	}
	defer db.Close()

	st, err := bolt.NewStorage(db)
	if err != nil {
		return err
	}

	// Quick setup if database doesn't exist
	if !databaseExisted {
		if err := quickSetup(st); err != nil {
			return err
		}
	}

	// Build image service
	if *imageProcessors < 1 {
		return errors.New("image resize workers count could not be < 1")
	}
	imageService := files.NewImage(*imageProcessors)

	// Setup file cache
	var fileCache files.Interface = files.NewNoOp()
	if *cacheDir != "" {
		if err := os.MkdirAll(*cacheDir, 0700); err != nil {
			return fmt.Errorf("can't make directory %s: %w", *cacheDir, err)
		}
		fileCache = files.New(afero.NewOsFs(), *cacheDir)
	}

	// Get server settings
	server, err := getServerSettings(st)
	if err != nil {
		return err
	}
	setupLog(server.Log)

	rootPath, err := filepath.Abs(server.Root)
	if err != nil {
		return err
	}
	server.Root = rootPath

	// Get and setup settings
	set, err := st.Settings.Get()
	if err != nil {
		return err
	}

	if err := setTOTPEncryptionKey(set, st); err != nil {
		return err
	}

	// Create listener
	adr := server.Address + ":" + server.Port
	var listener net.Listener

	switch {
	case server.TLSKey != "" && server.TLSCert != "":
		cer, err := tls.LoadX509KeyPair(server.TLSCert, server.TLSKey)
		if err != nil {
			return err
		}
		listener, err = tls.Listen("tcp", adr, &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cer}},
		)
		if err != nil {
			return err
		}
	default:
		listener, err = net.Listen("tcp", adr)
		if err != nil {
			return err
		}
	}

	// Create HTTP handler
	assetsFs, err := fs.Sub(www.Assets(), "dist")
	if err != nil {
		return err
	}

	handler, err := fbhttp.NewHandler(imageService, fileCache, st, server, assetsFs)
	if err != nil {
		return err
	}

	defer listener.Close()

	log.Println("Listening on", listener.Addr().String())
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	// Wait for signal
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	sig := <-sigc
	log.Println("Got signal:", sig)

	// Graceful shutdown
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")

	return nil
}

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply config values only if not already set by flags
	if cfg.Database != "" && !isFlagSet("database") {
		*database = cfg.Database
	}
	if cfg.Address != "" && !isFlagSet("address") {
		*address = cfg.Address
	}
	if cfg.Port != "" && !isFlagSet("port") {
		*port = cfg.Port
	}
	if cfg.Cert != "" && !isFlagSet("cert") {
		*cert = cfg.Cert
	}
	if cfg.Key != "" && !isFlagSet("key") {
		*key = cfg.Key
	}
	if cfg.Root != "" && !isFlagSet("root") {
		*root = cfg.Root
	}
	if cfg.BaseURL != "" && !isFlagSet("baseURL") {
		*baseURL = cfg.BaseURL
	}
	if cfg.Log != "" && !isFlagSet("log") {
		*logPath = cfg.Log
	}
	if cfg.CacheDir != "" && !isFlagSet("cacheDir") {
		*cacheDir = cfg.CacheDir
	}
	if cfg.TokenExpirationTime != "" && !isFlagSet("tokenExpirationTime") {
		*tokenExpirationTime = cfg.TokenExpirationTime
	}
	if cfg.TotpTokenExpirationTime != "" && !isFlagSet("totpTokenExpirationTime") {
		*totpTokenExpirationTime = cfg.TotpTokenExpirationTime
	}
	if cfg.DisableThumbnails != nil && !isFlagSet("disableThumbnails") {
		*disableThumbnails = *cfg.DisableThumbnails
	}
	if cfg.DisablePreviewResize != nil && !isFlagSet("disablePreviewResize") {
		*disablePreviewResize = *cfg.DisablePreviewResize
	}
	if cfg.DisableTypeDetectionByHeader != nil && !isFlagSet("disableTypeDetectionByHeader") {
		*disableTypeDetectionByHeader = *cfg.DisableTypeDetectionByHeader
	}
	if cfg.DisableTOTP != nil && !isFlagSet("disableTOTP") {
		*disableTOTP = *cfg.DisableTOTP
	}
	if cfg.DisablePasskey != nil && !isFlagSet("disablePasskey") {
		*disablePasskey = *cfg.DisablePasskey
	}
	if cfg.Noauth != nil && !isFlagSet("noauth") {
		*noauth = *cfg.Noauth
	}
	if cfg.Username != "" && !isFlagSet("username") {
		*username = cfg.Username
	}
	if cfg.Password != "" && !isFlagSet("password") {
		*password = cfg.Password
	}
	if cfg.ImageProcessors != nil && !isFlagSet("imageProcessors") {
		*imageProcessors = *cfg.ImageProcessors
	}

	return nil
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func dbExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.Size() != 0, nil
	}

	if os.IsNotExist(err) {
		d := filepath.Dir(path)
		_, err = os.Stat(d)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0700); err != nil {
				return false, err
			}
			return false, nil
		}
	}

	return false, err
}

func getServerSettings(st *storage.Storage) (*settings.Server, error) {
	server, err := st.Settings.GetServer()
	if err != nil {
		return nil, err
	}

	// Apply flag values
	server.Address = *address
	server.Port = *port
	server.TLSCert = *cert
	server.TLSKey = *key
	server.Root = *root
	server.BaseURL = *baseURL
	server.Log = *logPath
	server.TokenExpirationTime = *tokenExpirationTime
	server.TOTPTokenExpirationTime = *totpTokenExpirationTime
	server.EnableThumbnails = !*disableThumbnails
	server.ResizePreview = !*disablePreviewResize
	server.TypeDetectionByHeader = !*disableTypeDetectionByHeader
	server.EnableTOTP = !*disableTOTP
	server.EnablePasskey = !*disablePasskey

	return server, nil
}

func setupLog(logMethod string) {
	switch logMethod {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(io.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logMethod,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}
}

func quickSetup(st *storage.Storage) error {
	log.Println("Performing quick setup")

	set := &settings.Settings{
		Key:                   generateKey(),
		Signup:                false,
		HideLoginButton:       true,
		CreateUserDir:         false,
		MinimumPasswordLength: settings.DefaultMinimumPasswordLength,
		UserHomeBasePath:      settings.DefaultUsersHomeBasePath,
		Defaults: settings.UserDefaults{
			Scope:       ".",
			Locale:      "en",
			SingleClick: false,
			Perm: users.Permissions{
				Admin:    false,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
			TOTPEnabled:    true,
			PasskeyEnabled: true,
		},
		AuthMethod: "",
		Branding:   settings.Branding{},
		Tus: settings.Tus{
			ChunkSize:  settings.DefaultTusChunkSize,
			RetryCount: settings.DefaultTusRetryCount,
		},
		Shell:          nil,
		TOTPEnabled:    true,
		PasskeyEnabled: true,
	}

	var err error
	if *noauth {
		set.AuthMethod = auth.MethodNoAuth
		err = st.Auth.Save(&auth.NoAuth{})
	} else {
		set.AuthMethod = auth.MethodJSONAuth
		err = st.Auth.Save(&auth.JSONAuth{})
	}
	if err != nil {
		return err
	}

	if err = st.Settings.Save(set); err != nil {
		return err
	}

	// Save server settings
	ser := &settings.Server{
		BaseURL:                 *baseURL,
		Port:                    *port,
		Log:                     *logPath,
		TLSKey:                  *key,
		TLSCert:                 *cert,
		Address:                 *address,
		Root:                    *root,
		TokenExpirationTime:     *tokenExpirationTime,
		TOTPTokenExpirationTime: *totpTokenExpirationTime,
		EnableThumbnails:        !*disableThumbnails,
		ResizePreview:           !*disablePreviewResize,
		TypeDetectionByHeader:   !*disableTypeDetectionByHeader,
		EnableTOTP:              !*disableTOTP,
		EnablePasskey:           !*disablePasskey,
	}

	if err = st.Settings.SaveServer(ser); err != nil {
		return err
	}

	// Create admin user
	pwd := *password
	if pwd == "" {
		var genPwd string
		genPwd, err = users.RandomPwd(set.MinimumPasswordLength)
		if err != nil {
			return err
		}
		log.Printf("User '%s' initialized with randomly generated password: %s\n", *username, genPwd)
		pwd, err = users.ValidateAndHashPwd(genPwd, set.MinimumPasswordLength)
		if err != nil {
			return err
		}
	} else {
		log.Printf("User '%s' initialized with user-provided password\n", *username)
	}

	if *username == "" || pwd == "" {
		log.Fatal("username and password cannot be empty during quick setup")
	}

	user := &users.User{
		Username:     *username,
		Password:     pwd,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	return st.Users.Save(user)
}

func setTOTPEncryptionKey(set *settings.Settings, store *storage.Storage) error {
	// If key is already set (len 32), use it
	if len(set.TOTPEncryptionKey) == 32 {
		return nil
	}

	// Generate one and save it
	newKey, err := settings.GenerateKey()
	if err != nil {
		return err
	}
	// GenerateKey returns 64 bytes, we need 32 for AES-256
	set.TOTPEncryptionKey = newKey[:32]

	return store.Settings.Save(set)
}

func generateKey() []byte {
	k, err := settings.GenerateKey()
	if err != nil {
		panic(err)
	}
	return k
}
