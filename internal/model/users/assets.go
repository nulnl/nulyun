package users

import (
	"embed"
	"strings"
)

//go:embed assets
var assets embed.FS
var commonPasswords map[string]struct{}

func init() {
	data, err := assets.ReadFile("assets/common-passwords.txt")
	if err != nil {
		panic(err)
	}

	passwords := strings.Split(strings.TrimSpace(string(data)), "\n")
	commonPasswords = make(map[string]struct{}, len(passwords))
	for _, password := range passwords {
		commonPasswords[strings.TrimSpace(password)] = struct{}{}
	}
}
