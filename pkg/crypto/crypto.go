// Package crypto provides RSA encryption and decryption functionality.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// PublicKey represents an RSA public key
type PublicKey struct {
	key *rsa.PublicKey
}

// PrivateKey represents an RSA private key
type PrivateKey struct {
	key *rsa.PrivateKey
}

// LoadPublicKey loads an RSA public key from a PEM file
func LoadPublicKey(filename string) (*PublicKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA public key")
	}

	return &PublicKey{key: rsaPubKey}, nil
}

// LoadPrivateKey loads an RSA private key from a PEM file
func LoadPrivateKey(filename string) (*PrivateKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA private key")
	}

	return &PrivateKey{key: rsaPrivKey}, nil
}

// Encrypt encrypts data using hybrid encryption:
// 1. Generate a random AES-256 key
// 2. Encrypt the data with AES-GCM
// 3. Encrypt the AES key with RSA-OAEP
// Returns: [12-byte nonce][encrypted AES key length (2 bytes)][encrypted AES key][ciphertext]
func (p *PublicKey) Encrypt(data []byte) ([]byte, error) {
	// Generate random AES key
	aesKey := make([]byte, 32) // 256 bits
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Use GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data with AES-GCM
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Encrypt AES key with RSA
	encryptedAESKey, err := rsa.EncryptOAEP(
		nil,
		rand.Reader,
		p.key,
		aesKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// Combine: [nonce (12 bytes)][encrypted AES key length (2 bytes)][encrypted AES key][ciphertext]
	result := make([]byte, 0, 12+2+len(encryptedAESKey)+len(ciphertext))
	result = append(result, nonce...)
	// Add length of encrypted AES key as 2 bytes
	result = append(result, byte(len(encryptedAESKey)>>8), byte(len(encryptedAESKey)&0xFF))
	result = append(result, encryptedAESKey...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts data using hybrid decryption:
// 1. Extract and decrypt the AES key with RSA
// 2. Decrypt the data with AES-GCM
func (p *PrivateKey) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 12+2 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce
	nonce := ciphertext[:12]

	// Extract encrypted AES key length
	keyLen := int(ciphertext[12])<<8 | int(ciphertext[13])

	if len(ciphertext) < 14+keyLen {
		return nil, fmt.Errorf("ciphertext too short for encrypted key")
	}

	// Extract encrypted AES key
	encryptedAESKey := ciphertext[14 : 14+keyLen]

	// Extract actual encrypted data
	data := ciphertext[14+keyLen:]

	// Decrypt AES key with RSA
	aesKey, err := rsa.DecryptOAEP(
		nil,
		rand.Reader,
		p.key,
		encryptedAESKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
