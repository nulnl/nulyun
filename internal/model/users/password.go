package users

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	fberrors "github.com/nulnl/nulyun/internal/pkg_errors"
)

// ValidateAndHashPwd validates and hashes a password.
func ValidateAndHashPwd(password string, minimumLength uint) (string, error) {
	if uint(len(password)) < minimumLength {
		return "", fberrors.ErrShortPassword{MinimumLength: minimumLength}
	}

	if _, ok := commonPasswords[password]; ok {
		return "", fberrors.ErrEasyPassword
	}

	return HashPwd(password)
}

// HashPwd hashes a password.
func HashPwd(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPwd checks if a password is correct.
func CheckPwd(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func RandomPwd(passwordLength uint) (string, error) {
	randomPasswordBytes := make([]byte, passwordLength)
	var _, err = rand.Read(randomPasswordBytes)
	if err != nil {
		return "", err
	}

	// This is done purely to make the password human-readable
	var randomPasswordString = base64.URLEncoding.EncodeToString(randomPasswordBytes)
	return randomPasswordString, nil
}

// returns cipher text and nonce in base64
func EncryptSymmetric(encryptionKey, secret []byte) (string, string, error) {
	if len(encryptionKey) != 32 {
		log.Printf("%s (key=\"%s\")", fberrors.ErrInvalidEncryptionKey.Error(), string(encryptionKey))
		return "", "", fberrors.ErrInvalidEncryptionKey
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}

	cipherText := gcm.Seal(nil, nonce, secret, nil)

	return base64.StdEncoding.EncodeToString(cipherText), base64.StdEncoding.EncodeToString(nonce), nil
}

func DecryptSymmetric(encryptionKey []byte, cipherTextB64, nonceB64 string) (string, error) {
	if len(encryptionKey) != 32 {
		log.Printf("%s (key=\"%s\")", fberrors.ErrInvalidEncryptionKey.Error(), string(encryptionKey))
		return "", fberrors.ErrInvalidEncryptionKey
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	secret, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(secret), nil
}

// Decrypt the secret and validate the code
func CheckTOTP(totpEncryptionKey []byte, encryptedSecretB64, nonceB64, code string) (bool, error) {
	if len(totpEncryptionKey) != 32 {
		log.Printf("%s (key=\"%s\")", fberrors.ErrInvalidEncryptionKey.Error(), string(totpEncryptionKey))
		return false, fberrors.ErrInvalidEncryptionKey
	}

	secret, err := DecryptSymmetric(totpEncryptionKey, encryptedSecretB64, nonceB64)
	if err != nil {
		return false, err
	}

	return totp.Validate(code, secret), nil
}

// GenerateRecoveryCodes generates 10 recovery codes
func GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		// Generate 8 bytes of random data, encode to base32 (becomes ~13 chars)
		randomBytes := make([]byte, 8)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, err
		}
		// Format as XXXX-XXXX-XXXX for readability
		code := base64.RawStdEncoding.EncodeToString(randomBytes)
		// Take first 12 chars and format
		if len(code) > 12 {
			code = code[:12]
		}
		formattedCode := code[:4] + "-" + code[4:8] + "-" + code[8:]
		codes[i], err = HashPwd(formattedCode)
		if err != nil {
			return nil, err
		}
	}
	return codes, nil
}

// ValidateRecoveryCode checks if the provided code matches any unused recovery code
// Returns the index of the matched code, or -1 if no match found
func ValidateRecoveryCode(code string, hashedCodes []string) int {
	for i, hashedCode := range hashedCodes {
		if hashedCode != "" && CheckPwd(code, hashedCode) {
			return i
		}
	}
	return -1
}
