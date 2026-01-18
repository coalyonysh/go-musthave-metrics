package handlers

import (
	"net"
	"net/http"
	"strings"
)

// getClientIP извлекает IP адрес клиента из запроса
func getClientIP(r *http.Request) string {
	// Проверяем заголовок X-Real-IP (для прокси)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Проверяем заголовок X-Forwarded-For
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// Берем первый IP из списка
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return ip
	}
	// Используем RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
