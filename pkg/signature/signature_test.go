package signature

import (
	"testing"
)

func TestCalculateHash_WithKey(t *testing.T) {
	data := []byte("test data")
	key := "test_key"

	hash := CalculateHash(data, key)

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	// Same input should produce same hash
	hash2 := CalculateHash(data, key)
	if hash != hash2 {
		t.Errorf("expected same hash for same input, got %s and %s", hash, hash2)
	}
}

func TestCalculateHash_EmptyKey(t *testing.T) {
	data := []byte("test data")

	hash := CalculateHash(data, "")

	if hash != "" {
		t.Errorf("expected empty hash for empty key, got %s", hash)
	}
}

func TestCalculateHash_DifferentKeys(t *testing.T) {
	data := []byte("test data")

	hash1 := CalculateHash(data, "key1")
	hash2 := CalculateHash(data, "key2")

	if hash1 == hash2 {
		t.Error("expected different hashes for different keys")
	}
}

func TestCalculateHash_DifferentData(t *testing.T) {
	key := "test_key"

	hash1 := CalculateHash([]byte("data1"), key)
	hash2 := CalculateHash([]byte("data2"), key)

	if hash1 == hash2 {
		t.Error("expected different hashes for different data")
	}
}

func TestVerifyHash_ValidHash(t *testing.T) {
	data := []byte("test data")
	key := "test_key"

	computedHash := CalculateHash(data, key)
	result := VerifyHash(data, key, computedHash)

	if !result {
		t.Error("expected valid hash to pass verification")
	}
}

func TestVerifyHash_InvalidHash(t *testing.T) {
	data := []byte("test data")
	key := "test_key"

	result := VerifyHash(data, key, "invalid_hash")

	if result {
		t.Error("expected invalid hash to fail verification")
	}
}

func TestVerifyHash_EmptyKey(t *testing.T) {
	data := []byte("test data")

	// With empty key, VerifyHash should return true (skip verification)
	result := VerifyHash(data, "", "some_hash")

	if !result {
		t.Error("expected true when key is empty (skip verification)")
	}
}

func TestVerifyHash_EmptyProvidedHash(t *testing.T) {
	data := []byte("test data")
	key := "test_key"

	// With empty provided hash, VerifyHash should return true (skip verification)
	result := VerifyHash(data, key, "")

	if !result {
		t.Error("expected true when provided hash is empty (skip verification)")
	}
}

func TestVerifyHash_BothEmpty(t *testing.T) {
	data := []byte("test data")

	// Both key and providedHash empty - should return true
	result := VerifyHash(data, "", "")

	if !result {
		t.Error("expected true when both key and provided hash are empty")
	}
}

func TestVerifyHash_WrongKey(t *testing.T) {
	data := []byte("test data")
	correctKey := "correct_key"
	wrongKey := "wrong_key"

	computedHash := CalculateHash(data, correctKey)
	result := VerifyHash(data, wrongKey, computedHash)

	if result {
		t.Error("expected hash verification to fail with wrong key")
	}
}
