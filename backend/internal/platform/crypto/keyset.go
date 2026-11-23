package crypto

import (
	"errors"
	"fmt"
)

// KeySet supports rotation: new values use the current key, old values can be
// decrypted using any previously registered key ID.
type KeySet struct {
	current *Keyring
	byID    map[string]*Keyring
}

// NewKeySet registers one current key and any previous keys.
func NewKeySet(current *Keyring, previous ...*Keyring) (*KeySet, error) {
	if current == nil {
		return nil, errors.New("crypto: current keyring is nil")
	}
	set := &KeySet{current: current, byID: map[string]*Keyring{}}
	if err := set.add(current); err != nil {
		return nil, err
	}
	for _, k := range previous {
		if k == nil {
			return nil, errors.New("crypto: nil previous keyring")
		}
		if err := set.add(k); err != nil {
			return nil, err
		}
	}
	return set, nil
}

func (s *KeySet) add(k *Keyring) error {
	if _, ok := s.byID[k.KeyID()]; ok {
		return fmt.Errorf("crypto: duplicate key id %s", k.KeyID())
	}
	s.byID[k.KeyID()] = k
	return nil
}

func (s *KeySet) CurrentKeyID() string { return s.current.KeyID() }

func (s *KeySet) Encrypt(purpose string, aad AAD, plaintext []byte) ([]byte, string, error) {
	return s.current.Encrypt(purpose, aad, plaintext)
}

func (s *KeySet) Decrypt(purpose string, aad AAD, ciphertext []byte, keyID string) ([]byte, error) {
	k, ok := s.byID[keyID]
	if !ok {
		return nil, fmt.Errorf("crypto: unknown key id %s", keyID)
	}
	return k.Decrypt(purpose, aad, ciphertext, keyID)
}
