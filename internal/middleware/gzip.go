package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipCompress middleware для сжатия ответов
func GzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем wrapper для ResponseWriter, чтобы перехватить Content-Type
		gzWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         nil,
		}

		next.ServeHTTP(gzWriter, r)

		// Закрываем gzip writer, если он был создан
		if gzWriter.Writer != nil {
			gzWriter.Writer.Close()
		}
	})
}

// GzipDecompress middleware для распаковки запросов
func GzipDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, сжат ли запрос
		if r.Header.Get("Content-Encoding") == "gzip" {
			// Распаковываем тело запроса
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip data", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()

			// Заменяем тело запроса на распакованное
			r.Body = io.NopCloser(gzReader)
		}

		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter обертка для ResponseWriter с поддержкой gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer      *gzip.Writer
	contentType string
	wroteHeader bool
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Получаем Content-Type из заголовков (может быть установлен до WriteHeader)
	w.contentType = w.Header().Get("Content-Type")

	// Если Content-Type подходит для сжатия, создаем gzip writer
	if shouldCompress(w.contentType) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Удаляем Content-Length, так как размер изменится
		w.Writer = gzip.NewWriter(w.ResponseWriter)
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// Если Content-Type еще не установлен, проверяем его при первом Write
		if w.contentType == "" {
			w.contentType = w.Header().Get("Content-Type")
		}
		// Устанавливаем заголовки перед WriteHeader
		if shouldCompress(w.contentType) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
		}
		w.WriteHeader(http.StatusOK)
		// Создаем gzip writer после WriteHeader, если нужно
		if shouldCompress(w.contentType) && w.Writer == nil {
			w.Writer = gzip.NewWriter(w.ResponseWriter)
		}
	}

	if w.Writer != nil {
		return w.Writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// shouldCompress проверяет, нужно ли сжимать контент данного типа
func shouldCompress(contentType string) bool {
	// Сжимаем только application/json и text/html
	return strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/html")
}

