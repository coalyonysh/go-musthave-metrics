package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// CalculateHash вычисляет SHA256 хеш от data с использованием key
func CalculateHash(data []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// VerifyHash проверяет, соответствует ли предоставленный hash вычисленному от data с key
func VerifyHash(data []byte, key, providedHash string) bool {
	if key == "" {
		return true // если ключ не задан, пропускаем проверку
	}
	computed := CalculateHash(data, key)
	return hmac.Equal([]byte(computed), []byte(providedHash))
}
