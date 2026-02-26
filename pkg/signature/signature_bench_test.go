package signature

import (
	"testing"
)

func BenchmarkCalculateHash(b *testing.B) {
	data := []byte(`{"id":"test","type":"gauge","value":123.45}`)
	key := "test_secret_key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, key)
	}
}

func BenchmarkCalculateHash_EmptyKey(b *testing.B) {
	data := []byte(`{"id":"test","type":"gauge","value":123.45}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, "")
	}
}

func BenchmarkVerifyHash(b *testing.B) {
	data := []byte(`{"id":"test","type":"gauge","value":123.45}`)
	key := "test_secret_key"
	hash := CalculateHash(data, key)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyHash(data, key, hash)
	}
}

func BenchmarkVerifyHash_Invalid(b *testing.B) {
	data := []byte(`{"id":"test","type":"gauge","value":123.45}`)
	key := "test_secret_key"
	invalidHash := "invalid_hash"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyHash(data, key, invalidHash)
	}
}

func BenchmarkVerifyHash_EmptyKey(b *testing.B) {
	data := []byte(`{"id":"test","type":"gauge","value":123.45}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyHash(data, "", "some_hash")
	}
}

func BenchmarkCalculateHash_LargeData(b *testing.B) {
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	key := "test_secret_key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, key)
	}
}
