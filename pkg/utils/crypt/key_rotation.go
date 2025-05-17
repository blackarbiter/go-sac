package crypt

import (
	"crypto/rand"
	"errors"
	"sync"
	"time"
)

type KeyManager struct {
	currentKey     []byte
	previousKeys   [][]byte
	mu             sync.RWMutex
	rotationPeriod time.Duration
	lastRotated    time.Time
}

func NewKeyManager(initialKey []byte, rotationPeriod time.Duration) *KeyManager {
	return &KeyManager{
		currentKey:     initialKey,
		rotationPeriod: rotationPeriod,
		lastRotated:    time.Now(),
	}
}

// StartAutoRotation 启动自动密钥轮换
func (km *KeyManager) StartAutoRotation() {
	go func() {
		for {
			time.Sleep(km.rotationPeriod)
			km.rotateKey()
		}
	}()
}

func (km *KeyManager) rotateKey() {
	km.mu.Lock()
	defer km.mu.Unlock()

	newKey := make([]byte, 32) // AES-256
	if _, err := rand.Read(newKey); err != nil {
		panic("failed to generate new key: " + err.Error())
	}

	km.previousKeys = append([][]byte{km.currentKey}, km.previousKeys...)
	if len(km.previousKeys) > 3 { // 保留最近3个旧密钥
		km.previousKeys = km.previousKeys[:3]
	}

	km.currentKey = newKey
	km.lastRotated = time.Now()
}

// DecryptWithHistory 尝试用当前和历史密钥解密
func (km *KeyManager) DecryptWithHistory(ciphertext []byte) ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	keys := append([][]byte{km.currentKey}, km.previousKeys...)
	for _, key := range keys {
		if plaintext, err := Decrypt(ciphertext, key); err == nil {
			return plaintext, nil
		}
	}
	return nil, errors.New("decryption failed with all keys")
}
