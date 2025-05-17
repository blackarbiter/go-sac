package crypt_test

import (
	"crypto/rand"
	"testing"

	"github.com/blackarbiter/go-sac/pkg/utils/crypt"
)

func TestAESGCM(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	original := []byte("sensitive data")

	encrypted, err := crypt.Encrypt(original, key)
	if err != nil {
		t.Fatal("Encryption failed:", err)
	}

	decrypted, err := crypt.Decrypt(encrypted, key)
	if err != nil {
		t.Fatal("Decryption failed:", err)
	}

	if string(decrypted) != string(original) {
		t.Error("Decrypted text mismatch")
	}
}
