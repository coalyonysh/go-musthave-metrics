package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestLoadPublicKey_FileNotFound(t *testing.T) {
	_, err := LoadPublicKey("nonexistent_key.pem")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadPrivateKey_FileNotFound(t *testing.T) {
	_, err := LoadPrivateKey("nonexistent_key.pem")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pubKey := &PublicKey{key: &privateKey.PublicKey}
	privKey := &PrivateKey{key: privateKey}

	// Test data
	plaintext := []byte("Hello, World! This is a test message.")

	// Encrypt
	ciphertext, err := pubKey.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt
	decrypted, err := privKey.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Verify
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted text doesn't match original: got %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestEncryptEmptyData(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pubKey := &PublicKey{key: &privateKey.PublicKey}

	// Test with empty data
	ciphertext, err := pubKey.Encrypt([]byte{})
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	privKey := &PrivateKey{key: privateKey}
	decrypted, err := privKey.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty decrypted data, got %d bytes", len(decrypted))
	}
}

func TestEncryptLargeData(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pubKey := &PublicKey{key: &privateKey.PublicKey}

	// Test with larger data (1MB)
	plaintext := make([]byte, 1024*1024)
	rand.Read(plaintext)

	ciphertext, err := pubKey.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	privKey := &PrivateKey{key: privateKey}
	decrypted, err := privKey.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if len(decrypted) != len(plaintext) {
		t.Errorf("decrypted length doesn't match: got %d, want %d", len(decrypted), len(plaintext))
	}
}

func TestDecryptInvalidCiphertextTooShort(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	privKey := &PrivateKey{key: privateKey}

	// Test with too short ciphertext
	_, err = privKey.Decrypt([]byte("short"))
	if err == nil {
		t.Error("expected error for too short ciphertext")
	}
}

func TestDecryptInvalidKeyLen(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	privKey := &PrivateKey{key: privateKey}

	// Create a ciphertext with invalid key length indicator (but proper nonce)
	// nonce (12) + key len (2) + invalid key + data
	ciphertext := make([]byte, 14)
	ciphertext[12] = 0xFF // Set key length to 255 (more than available)
	ciphertext[13] = 0xFF

	_, err = privKey.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error for invalid key length")
	}
}

func TestPublicKeyEncryptNil(t *testing.T) {
	pubKey := &PublicKey{key: nil}
	_, err := pubKey.Encrypt([]byte("test"))
	if err == nil {
		t.Error("expected error for nil key")
	}
}

func TestPrivateKeyDecryptNil(t *testing.T) {
	privKey := &PrivateKey{key: nil}
	_, err := privKey.Decrypt([]byte("test data"))
	if err == nil {
		t.Error("expected error for nil key")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	// Generate two different key pairs
	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	key2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pubKey := &PublicKey{key: &key1.PublicKey}
	privKeyWrong := &PrivateKey{key: key2}

	plaintext := []byte("test message")
	ciphertext, err := pubKey.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decryption with wrong key should fail
	_, err = privKeyWrong.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error for decryption with wrong key")
	}
}
