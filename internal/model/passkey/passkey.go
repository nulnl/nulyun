package passkey

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Credential represents a WebAuthn credential for a user
type Credential struct {
	ID              uint          `storm:"id,increment" json:"id"`
	UserID          uint          `json:"userId"`
	Name            string        `json:"name"`            // User-friendly name for the credential
	CredentialID    []byte        `json:"credentialId"`    // WebAuthn credential ID
	PublicKey       []byte        `json:"publicKey"`       // Public key
	AttestationType string        `json:"attestationType"` // Attestation type
	Transport       []string      `json:"transport"`       // Transport methods
	Flags           Flags         `json:"flags"`           // Authenticator flags
	Authenticator   Authenticator `json:"authenticator"`   // Authenticator info
	CreatedAt       time.Time     `json:"createdAt"`
	LastUsedAt      time.Time     `json:"lastUsedAt"`
}

// Flags represents authenticator flags
type Flags struct {
	UserPresent    bool `json:"userPresent"`
	UserVerified   bool `json:"userVerified"`
	BackupEligible bool `json:"backupEligible"`
	BackupState    bool `json:"backupState"`
}

// Authenticator represents authenticator information
type Authenticator struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"signCount"`
	CloneWarning bool   `json:"cloneWarning"`
}

// ToWebAuthnCredential converts to webauthn.Credential
func (c *Credential) ToWebAuthnCredential() webauthn.Credential {
	transport := make([]protocol.AuthenticatorTransport, len(c.Transport))
	for i, t := range c.Transport {
		transport[i] = protocol.AuthenticatorTransport(t)
	}

	return webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transport,
		Flags: webauthn.CredentialFlags{
			UserPresent:    c.Flags.UserPresent,
			UserVerified:   c.Flags.UserVerified,
			BackupEligible: c.Flags.BackupEligible,
			BackupState:    c.Flags.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:       c.Authenticator.AAGUID,
			SignCount:    c.Authenticator.SignCount,
			CloneWarning: c.Authenticator.CloneWarning,
		},
	}
}

// FromWebAuthnCredential creates a Credential from webauthn.Credential
func FromWebAuthnCredential(userID uint, name string, cred webauthn.Credential) *Credential {
	transport := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transport[i] = string(t)
	}

	return &Credential{
		UserID:          userID,
		Name:            name,
		CredentialID:    cred.ID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		Transport:       transport,
		Flags: Flags{
			UserPresent:    cred.Flags.UserPresent,
			UserVerified:   cred.Flags.UserVerified,
			BackupEligible: cred.Flags.BackupEligible,
			BackupState:    cred.Flags.BackupState,
		},
		Authenticator: Authenticator{
			AAGUID:       cred.Authenticator.AAGUID,
			SignCount:    cred.Authenticator.SignCount,
			CloneWarning: cred.Authenticator.CloneWarning,
		},
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
}

// Store defines the interface for passkey storage
type Store interface {
	// Save saves a credential
	Save(cred *Credential) error
	// Get gets a credential by ID
	Get(id uint) (*Credential, error)
	// GetByUserID gets all credentials for a user
	GetByUserID(userID uint) ([]*Credential, error)
	// GetByCredentialID gets a credential by its credential ID
	GetByCredentialID(credentialID []byte) (*Credential, error)
	// Update updates a credential
	Update(cred *Credential, fields ...string) error
	// Delete deletes a credential
	Delete(id uint) error
}
