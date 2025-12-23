package fbhttp

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	fberrors "github.com/nulnl/nulyun/internal/errors"
	"github.com/nulnl/nulyun/internal/settings/users"
)

var (
	NonModifiableFieldsForNonAdmin = []string{"Username", "Scope", "LockPassword", "Perm"}
	TOTPIssuer                     = "nulyun"
)

type modifyUserRequest struct {
	modifyRequest
	Data *users.User `json:"data"`
}

type enableTOTPVerificationRequest struct {
	Password string `json:"password"`
}

type enableTOTPVerificationResponse struct {
	SetupKey string `json:"setupKey"`
}

type getTOTPInfoResponse struct {
	SetupKey string `json:"setupKey"`
}

type checkTOTPRequest struct {
	Code string `json:"code"`
}

func getUserID(r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	i, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), err
}

func getUser(_ http.ResponseWriter, r *http.Request) (*modifyUserRequest, error) {
	if r.Body == nil {
		return nil, fberrors.ErrEmptyRequest
	}

	req := &modifyUserRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return nil, err
	}

	if req.What != "user" {
		return nil, fberrors.ErrInvalidDataType
	}

	return req, nil
}

func withSelfOrAdmin(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		id, err := getUserID(r)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if d.user.ID != id && !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}

		d.raw = id
		return fn(w, r, d)
	})
}

var usersGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	users, err := d.store.Users.Gets(d.server.Root)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	for _, u := range users {
		u.Password = ""
		u.TOTPSecret = ""
		u.TOTPNonce = ""
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return renderJSON(w, r, users)
})

var userGetHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	u, err := d.store.Users.Get(d.server.Root, d.raw.(uint))
	if errors.Is(err, fberrors.ErrNotExist) {
		return http.StatusNotFound, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	u.Password = ""
	u.TOTPSecret = ""
	u.TOTPNonce = ""
	if !d.user.Perm.Admin {
		u.Scope = ""
	}
	return renderJSON(w, r, u)
})

var userDeleteHandler = withSelfOrAdmin(func(_ http.ResponseWriter, _ *http.Request, d *data) (int, error) {
	err := d.store.Users.Delete(d.raw.(uint))
	if err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})

var userPostHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req, err := getUser(w, r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if len(req.Which) != 0 {
		return http.StatusBadRequest, nil
	}

	if req.Data.Password == "" {
		return http.StatusBadRequest, fberrors.ErrEmptyPassword
	}

	req.Data.Password, err = users.ValidateAndHashPwd(req.Data.Password, d.settings.MinimumPasswordLength)
	if err != nil {
		return http.StatusBadRequest, err
	}

	userHome, err := d.settings.MakeUserDir(req.Data.Username, req.Data.Scope, d.server.Root)
	if err != nil {
		log.Printf("create user: failed to mkdir user home dir: [%s]", userHome)
		return http.StatusInternalServerError, err
	}
	req.Data.Scope = userHome
	log.Printf("user: %s, home dir: [%s].", req.Data.Username, userHome)

	err = d.store.Users.Save(req.Data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Location", "/settings/users/"+strconv.FormatUint(uint64(req.Data.ID), 10))
	return http.StatusCreated, nil
})

var userPutHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req, err := getUser(w, r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if req.Data.ID != d.raw.(uint) {
		return http.StatusBadRequest, nil
	}

	if len(req.Which) == 0 || (len(req.Which) == 1 && req.Which[0] == "all") {
		if !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}

		if req.Data.Password != "" {
			req.Data.Password, err = users.ValidateAndHashPwd(req.Data.Password, d.settings.MinimumPasswordLength)
			if err != nil {
				return http.StatusBadRequest, err
			}
		} else {
			var suser *users.User
			suser, err = d.store.Users.Get(d.server.Root, d.raw.(uint))
			if err != nil {
				return http.StatusInternalServerError, err
			}
			req.Data.Password = suser.Password
		}

		req.Which = []string{}
	}

	for k, v := range req.Which {
		v = cases.Title(language.English, cases.NoLower).String(v)
		req.Which[k] = v

		if v == "Password" {
			if !d.user.Perm.Admin && d.user.LockPassword {
				return http.StatusForbidden, nil
			}

			req.Data.Password, err = users.ValidateAndHashPwd(req.Data.Password, d.settings.MinimumPasswordLength)
			if err != nil {
				return http.StatusBadRequest, err
			}
		}

		for _, f := range NonModifiableFieldsForNonAdmin {
			if !d.user.Perm.Admin && v == f {
				return http.StatusForbidden, nil
			}
		}
	}

	err = d.store.Users.Update(req.Data, req.Which...)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})

var userEnableTOTPHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if r.Body == nil {
		return http.StatusBadRequest, fberrors.ErrEmptyRequest
	}

	if d.user.TOTPSecret != "" {
		return http.StatusBadRequest, fmt.Errorf("TOTP verification already enabled")
	}

	var req enableTOTPVerificationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Invalid request body: %w", err)
	} else if req.Password == "" {
		return http.StatusBadRequest, fberrors.ErrEmptyPassword
	} else if !users.CheckPwd(req.Password, d.user.Password) {
		return http.StatusBadRequest, errors.New("password is incorrect")
	}

	ops := totp.GenerateOpts{AccountName: d.user.Username, Issuer: TOTPIssuer}
	key, err := totp.Generate(ops)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	encryptedSecret, nonce, err := users.EncryptSymmetric(d.settings.TOTPEncryptionKey, []byte(key.Secret()))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	d.user.TOTPSecret = encryptedSecret
	d.user.TOTPNonce = nonce
	if err := d.store.Users.Update(d.user, "TOTPSecret", "TOTPNonce"); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, enableTOTPVerificationResponse{SetupKey: key.URL()})
})

var userGetTOTPHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.user.TOTPSecret == "" {
		return http.StatusForbidden, fmt.Errorf("user does not enable the TOTP verification")
	}

	// If TOTP is already verified, require TOTP code for security
	if d.user.TOTPVerified {
		code := r.Header.Get("X-TOTP-CODE")
		if code == "" {
			return http.StatusForbidden, nil
		}
		if ok, err := users.CheckTOTP(d.settings.TOTPEncryptionKey, d.user.TOTPSecret, d.user.TOTPNonce, code); err != nil {
			return http.StatusInternalServerError, err
		} else if !ok {
			return http.StatusForbidden, nil
		}
	}
	// If not verified yet, allow viewing the setup key without TOTP code

	secret, err := users.DecryptSymmetric(d.settings.TOTPEncryptionKey, d.user.TOTPSecret, d.user.TOTPNonce)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// DecryptSymmetric returns the Base32-encoded secret string (what we
	// originally got from key.Secret()). totp.Generate expects raw secret
	// bytes when passing the Secret option, so decode Base32 first to avoid
	// double-encoding (which produces a longer secret string).
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ops := totp.GenerateOpts{AccountName: d.user.Username, Issuer: TOTPIssuer, Secret: decoded}
	key, err := totp.Generate(ops)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, getTOTPInfoResponse{SetupKey: key.URL()})
})

var userDisableTOTPHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.user.TOTPSecret == "" {
		return http.StatusOK, nil
	}

	// If TOTP is already verified, require TOTP code for security
	if d.user.TOTPVerified {
		code := r.Header.Get("X-TOTP-CODE")
		if code == "" {
			return http.StatusForbidden, nil
		}
		if ok, err := users.CheckTOTP(d.settings.TOTPEncryptionKey, d.user.TOTPSecret, d.user.TOTPNonce, code); err != nil {
			return http.StatusInternalServerError, err
		} else if !ok {
			return http.StatusForbidden, nil
		}
	}
	// If not verified yet, allow disabling without TOTP code

	d.user.TOTPNonce = ""
	d.user.TOTPSecret = ""
	d.user.TOTPVerified = false

	if err := d.store.Users.Update(d.user, "TOTPSecret", "TOTPNonce", "TOTPVerified"); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})

var userCheckTOTPHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.user.TOTPSecret == "" {
		return http.StatusForbidden, nil
	}

	var req checkTOTPRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Invalid request body: %w", err)
	}

	if ok, err := users.CheckTOTP(d.settings.TOTPEncryptionKey, d.user.TOTPSecret, d.user.TOTPNonce, req.Code); err != nil {
		return http.StatusInternalServerError, err
	} else if !ok {
		return http.StatusForbidden, nil
	}

	// Mark TOTP as verified after successful verification
	if !d.user.TOTPVerified {
		d.user.TOTPVerified = true
		if err := d.store.Users.Update(d.user, "TOTPVerified"); err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
})

// userResetTOTPHandler resets TOTP secret for a user (self or admin)
var userResetTOTPHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	targetUser, err := d.store.Users.Get(d.server.Root, d.raw.(uint))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Verify password or TOTP for security
	if d.user.ID == targetUser.ID {
		// User resetting their own TOTP - require password
		if r.Body == nil {
			return http.StatusBadRequest, fberrors.ErrEmptyRequest
		}

		var req enableTOTPVerificationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("Invalid request body: %w", err)
		}
		if req.Password == "" {
			return http.StatusBadRequest, fberrors.ErrEmptyPassword
		}
		if !users.CheckPwd(req.Password, d.user.Password) {
			return http.StatusBadRequest, errors.New("password is incorrect")
		}
	}
	// Admin can reset without password

	// Generate new TOTP secret
	ops := totp.GenerateOpts{AccountName: targetUser.Username, Issuer: TOTPIssuer}
	key, err := totp.Generate(ops)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	encryptedSecret, nonce, err := users.EncryptSymmetric(d.settings.TOTPEncryptionKey, []byte(key.Secret()))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	targetUser.TOTPSecret = encryptedSecret
	targetUser.TOTPNonce = nonce
	targetUser.TOTPVerified = false
	targetUser.RecoveryCodes = []string{} // Clear old recovery codes

	if err := d.store.Users.Update(targetUser, "TOTPSecret", "TOTPNonce", "TOTPVerified", "RecoveryCodes"); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, enableTOTPVerificationResponse{SetupKey: key.URL()})
})

type recoveryCodesResponse struct {
	Codes []string `json:"codes"`
}

// userGenerateRecoveryCodesHandler generates new recovery codes for a user
var userGenerateRecoveryCodesHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	targetUser, err := d.store.Users.Get(d.server.Root, d.raw.(uint))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Check if TOTP is enabled for this user
	if targetUser.TOTPSecret == "" || !targetUser.TOTPVerified {
		return http.StatusBadRequest, fmt.Errorf("TOTP must be enabled and verified before generating recovery codes")
	}

	// Verify password or TOTP for security
	if d.user.ID == targetUser.ID {
		// User generating their own codes
		if targetUser.TOTPVerified {
			code := r.Header.Get("X-TOTP-CODE")
			if code == "" {
				return http.StatusForbidden, nil
			}
			if ok, err := users.CheckTOTP(d.settings.TOTPEncryptionKey, targetUser.TOTPSecret, targetUser.TOTPNonce, code); err != nil {
				return http.StatusInternalServerError, err
			} else if !ok {
				return http.StatusForbidden, nil
			}
		}
	}
	// Admin can generate without TOTP

	// Generate recovery codes
	codes, err := users.GenerateRecoveryCodes()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Store hashed codes
	targetUser.RecoveryCodes = codes
	if err := d.store.Users.Update(targetUser, "RecoveryCodes"); err != nil {
		return http.StatusInternalServerError, err
	}

	// Return plain text codes to user (only time they'll see them)
	// We need to regenerate the plain codes since we stored hashed versions
	newCodes, err := generatePlainRecoveryCodes()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Hash and update
	hashedCodes := make([]string, len(newCodes))
	for i, code := range newCodes {
		hashedCodes[i], err = users.HashPwd(code)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	targetUser.RecoveryCodes = hashedCodes
	if err := d.store.Users.Update(targetUser, "RecoveryCodes"); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, recoveryCodesResponse{Codes: newCodes})
})

// Helper function to generate plain recovery codes
func generatePlainRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		randomBytes := make([]byte, 8)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, err
		}
		code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
		if len(code) > 12 {
			code = code[:12]
		}
		// Format as XXXX-XXXX-XXXX
		formattedCode := code[:4] + "-" + code[4:8] + "-" + code[8:]
		codes[i] = formattedCode
	}
	return codes, nil
}

// userToggleTOTPHandler enables or disables TOTP for a user (admin only for other users)
var userToggleTOTPHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.server.EnableTOTP {
		return http.StatusForbidden, fmt.Errorf("TOTP feature is disabled")
	}

	targetUser, err := d.store.Users.Get(d.server.Root, d.raw.(uint))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type toggleRequest struct {
		Enabled bool `json:"enabled"`
	}

	var req toggleRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Invalid request body: %w", err)
	}

	targetUser.TOTPEnabled = req.Enabled
	if err := d.store.Users.Update(targetUser, "TOTPEnabled"); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})

// userTogglePasskeyHandler enables or disables Passkey for a user (admin only for other users)
var userTogglePasskeyHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.server.EnablePasskey {
		return http.StatusForbidden, fmt.Errorf("Passkey feature is disabled")
	}

	targetUser, err := d.store.Users.Get(d.server.Root, d.raw.(uint))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type toggleRequest struct {
		Enabled bool `json:"enabled"`
	}

	var req toggleRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Invalid request body: %w", err)
	}

	targetUser.PasskeyEnabled = req.Enabled
	if err := d.store.Users.Update(targetUser, "PasskeyEnabled"); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})
