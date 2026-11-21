package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params configures argon2id.
type Params struct {
	Memory     uint32
	Time       uint32
	Threads    uint8
	SaltLength uint32
	KeyLength  uint32
}

// DefaultArgon2Params are the production parameters.
//
//nolint:gochecknoglobals
var DefaultArgon2Params = Params{
	Memory:     19456,
	Time:       2,
	Threads:    1,
	SaltLength: 16,
	KeyLength:  32,
}

// Hasher hashes and verifies passwords.
type Hasher struct {
	params Params
	dummy  string
	sem    chan struct{}
}

// NewHasher creates a hasher.
func NewHasher(params Params) (*Hasher, error) {
	if params.Memory == 0 || params.Time == 0 || params.Threads == 0 || params.SaltLength == 0 || params.KeyLength == 0 {
		return nil, errors.New("identity: invalid argon2 parameters")
	}
	h := &Hasher{params: params, sem: make(chan struct{}, maxConcurrentHashes())}
	dummy, err := h.Hash("dummy-password-not-used")
	if err != nil {
		return nil, err
	}
	h.dummy = dummy
	return h, nil
}

func maxConcurrentHashes() int {
	n := runtime.NumCPU()
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4
	}
	return n
}

// Hash returns a PHC-encoded argon2id hash.
func (h *Hasher) Hash(password string) (string, error) {
	if len(password) > 1024 {
		return "", ErrPasswordTooLong
	}
	h.acquire()
	defer h.release()
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, h.params.Time, h.params.Memory, h.params.Threads, h.params.KeyLength)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.params.Memory, h.params.Time, h.params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// Verify checks a password against a PHC hash.
func (h *Hasher) Verify(encoded, password string) (bool, error) {
	h.acquire()
	defer h.release()
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, errors.New("identity: invalid password hash")
	}
	if !strings.HasPrefix(parts[2], "v=") {
		return false, errors.New("identity: invalid password hash")
	}
	if parts[2] != "v=19" {
		return false, errors.New("identity: unsupported argon2 version")
	}
	var memory uint32
	var timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false, errors.New("identity: invalid password hash")
	}
	if memory == 0 || memory > 1<<20 || timeCost == 0 || timeCost > 10 || threads == 0 || threads > 16 {
		return false, errors.New("identity: invalid password hash parameters")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errors.New("identity: invalid password hash")
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, errors.New("identity: invalid password hash")
	}
	key := argon2.IDKey([]byte(password), salt, timeCost, memory,
		threads, uint32(len(expected))) // #nosec G115
	return subtle.ConstantTimeCompare(key, expected) == 1, nil
}

// VerifyDummy performs a dummy verification.
func (h *Hasher) VerifyDummy(password string) error {
	_, err := h.Verify(h.dummy, password)
	return err
}

func (h *Hasher) acquire() {
	h.sem <- struct{}{}
}

func (h *Hasher) release() {
	<-h.sem
}
