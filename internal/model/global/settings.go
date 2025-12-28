package settings

import (
	"crypto/rand"
	"io/fs"
	"log"
	"strings"
	"time"
)

const DefaultUsersHomeBasePath = "/.users"
const DefaultLogoutPage = "/login"
const DefaultMinimumPasswordLength = 12
const DefaultFileMode = 0640
const DefaultDirMode = 0750

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key                   []byte       `json:"key"`
	Signup                bool         `json:"signup"`
	HideLoginButton       bool         `json:"hideLoginButton"`
	CreateUserDir         bool         `json:"createUserDir"`
	UserHomeBasePath      string       `json:"userHomeBasePath"`
	Defaults              UserDefaults `json:"defaults"`
	AuthMethod            AuthMethod   `json:"authMethod"`
	LogoutPage            string       `json:"logoutPage"`
	Branding              Branding     `json:"branding"`
	Tus                   Tus          `json:"tus"`
	Shell                 []string     `json:"shell"`
	MinimumPasswordLength uint         `json:"minimumPasswordLength"`
	FileMode              fs.FileMode  `json:"fileMode"`
	DirMode               fs.FileMode  `json:"dirMode"`
	HideDotfiles          bool         `json:"hideDotfiles"`
	TOTPEncryptionKey     []byte       `json:"totpEncryptionKey"`
	TOTPEnabled           bool         `json:"totpEnabled"`
}

// Server specific settings.
type Server struct {
	Root                    string `json:"root"`
	BaseURL                 string `json:"baseURL"`
	TLSKey                  string `json:"tlsKey"`
	TLSCert                 string `json:"tlsCert"`
	Port                    string `json:"port"`
	Address                 string `json:"address"`
	Log                     string `json:"log"`
	EnableThumbnails        bool   `json:"enableThumbnails"`
	ResizePreview           bool   `json:"resizePreview"`
	TypeDetectionByHeader   bool   `json:"typeDetectionByHeader"`
	AuthHook                string `json:"authHook"`
	TokenExpirationTime     string `json:"tokenExpirationTime"`
	TOTPTokenExpirationTime string `json:"totpTokenExpirationTime"`
	EnableTOTP              bool   `json:"enableTOTP"`
}

// Clean cleans any variables that might need cleaning.
func (s *Server) Clean() {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")
}

func (s *Server) GetTokenExpirationTime(tokenFB, totpFB time.Duration) (time.Duration, time.Duration) {
	getTokenDuration := func(v string, fb time.Duration) time.Duration {
		if v == "" {
			return fb
		}

		dur, err := time.ParseDuration(v)
		if err != nil {
			log.Printf("[WARN] Failed to parse ExpirationTime(value: %s): %v", v, err)
			return fb
		}
		return dur
	}

	return getTokenDuration(s.TokenExpirationTime, tokenFB), getTokenDuration(s.TOTPTokenExpirationTime, totpFB)
}

// GenerateKey generates a key of 512 bits.
func GenerateKey() ([]byte, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
