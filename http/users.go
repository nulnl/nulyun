package fbhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"encoding/base32"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	fberrors "github.com/nulnl/nulyun/errors"
	"github.com/nulnl/nulyun/settings/users"
)

var (
	NonModifiableFieldsForNonAdmin = []string{"Username", "Scope", "LockPassword", "Perm", "Commands", "Rules"}
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

var userGetTOTPHandler = withTOTP(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.user.TOTPSecret == "" {
		return http.StatusForbidden, fmt.Errorf("user does not enable the TOTP verification")
	}

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

	return renderJSON(w, r, getTOTPInfoResponse{SetupKey: key.URL()})
})

var userDisableTOTPHandler = withTOTP(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if d.user.TOTPSecret == "" {
		return http.StatusOK, nil
	}

	d.user.TOTPNonce = ""
	d.user.TOTPSecret = ""

	if err := d.store.Users.Update(d.user, "TOTPSecret", "TOTPNonce"); err != nil {
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

	return http.StatusOK, nil
})
