package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/pkg/crypto"
	"github.com/coalyonysh/go-musthave-metrics/pkg/signature"
)

type MetricSender interface {
	SendMetric(metric models.Metric) error
	SendMetricJSON(metric models.Metric) error
	SendMetricsBatch(metrics []models.Metric) error
}

type MetricHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	key        string
	cryptoKey  *crypto.PublicKey
	localIP    string
}

func NewMetricHTTPClient(baseURL string, key string) *MetricHTTPClient {
	localIP := getLocalIP()
	return &MetricHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		key:     key,
		localIP: localIP,
	}
}

func NewMetricHTTPClientWithCrypto(baseURL string, key string, cryptoKeyPath string) (*MetricHTTPClient, error) {
	localIP := getLocalIP()
	client := &MetricHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		key:     key,
		localIP: localIP,
	}

	if cryptoKeyPath != "" {
		pubKey, err := crypto.LoadPublicKey(cryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load crypto key: %w", err)
		}
		client.cryptoKey = pubKey
	}

	return client, nil
}

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	// Сначала попробуем подключиться к внешнему адресу, чтобы определить локальный IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Если не удалось, попробуем через UDP к любому локальному адресу
		conn, err = net.Dial("udp", "192.168.0.0:80")
		if err != nil {
			return "127.0.0.1"
		}
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// addXRealIPHeader adds X-Real-IP header to the request
func (c *MetricHTTPClient) addXRealIPHeader(req *http.Request) {
	if c.localIP != "" {
		req.Header.Set("X-Real-IP", c.localIP)
	}
}

// isRetriableError проверяет, является ли ошибка retriable
func isRetriableError(err error) bool {
	if err == nil {
		return false
	}
	// Сетевые ошибки
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	// Таймауты
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// HTTP ошибки, связанные с соединением
	var httpErr *url.Error
	return errors.As(err, &httpErr)
}

// retrySend выполняет функцию с повторными попытками
func retrySend(fn func() error) error {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for i, delay := range delays {
		err := fn()
		if err == nil {
			return nil
		}
		if !isRetriableError(err) {
			return err
		}
		log.Printf("Attempt %d failed, retrying in %v: %v", i+1, delay, err)
		time.Sleep(delay)
	}
	// Последняя попытка
	return fn()
}

func (c *MetricHTTPClient) SendMetric(metric models.Metric) error {
	var value string

	if metric.MType == models.Counter && metric.Delta != nil {
		value = fmt.Sprintf("%d", *metric.Delta)
	} else if metric.MType == models.Gauge && metric.Value != nil {
		value = fmt.Sprintf("%g", *metric.Value)
	} else {
		return fmt.Errorf("invalid metric: missing value for type %s", metric.MType)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s",
		c.baseURL, metric.MType, metric.ID, value)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	c.addXRealIPHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	log.Printf("Successfully sent metric: %s/%s/%s", metric.MType, metric.ID, value)
	return nil
}

// SendMetricJSON отправляет метрику в JSON формате через POST /update с gzip сжатием
func (c *MetricHTTPClient) SendMetricJSON(metric models.Metric) error {
	// Валидация метрики
	if metric.MType == models.Counter && metric.Delta == nil {
		return fmt.Errorf("invalid metric: missing delta for counter type")
	}
	if metric.MType == models.Gauge && metric.Value == nil {
		return fmt.Errorf("invalid metric: missing value for gauge type")
	}

	// Создаем JSON тело запроса
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric to JSON: %w", err)
	}

	// Сжимаем данные с помощью gzip
	var compressedData bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedData)
	if _, err := gzWriter.Write(jsonData); err != nil {
		gzWriter.Close()
		return fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return retrySend(func() error {
		url := fmt.Sprintf("%s/update", c.baseURL)

		req, err := http.NewRequest("POST", url, bytes.NewReader(compressedData.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		// Добавляем шифрование, если ключ задан
		if c.cryptoKey != nil {
			encryptedData, err := c.cryptoKey.Encrypt(compressedData.Bytes())
			if err != nil {
				return fmt.Errorf("failed to encrypt data: %w", err)
			}
			req.Body = nil
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(encryptedData)), nil
			}
			req.ContentLength = int64(len(encryptedData))
			req.Header.Set("Content-Encoding", "gzip") // keep gzip encoding header
			req.Header.Set("X-Encrypted", "true")
		}

		// Добавляем хеш, если ключ задан (до шифрования)
		if c.key != "" && c.cryptoKey == nil {
			hash := signature.CalculateHash(jsonData, c.key)
			req.Header.Set("HashSHA256", hash)
		}

		c.addXRealIPHeader(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Распаковываем ответ, если он сжат
		var responseBody io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			var gzReader *gzip.Reader
			gzReader, err = gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gzReader.Close()
			responseBody = gzReader
		}

		// Читаем ответ (для проверки, что все в порядке)
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		var valueStr string
		if metric.MType == models.Counter && metric.Delta != nil {
			valueStr = fmt.Sprintf("%d", *metric.Delta)
		} else if metric.MType == models.Gauge && metric.Value != nil {
			valueStr = fmt.Sprintf("%g", *metric.Value)
		}

		log.Printf("Successfully sent metric: %s/%s/%s", metric.MType, metric.ID, valueStr)
		return nil
	})
}

// SendMetricsBatch отправляет батч метрик в JSON формате через POST /updates с gzip сжатием
func (c *MetricHTTPClient) SendMetricsBatch(metrics []models.Metric) error {
	if len(metrics) == 0 {
		return nil // не отправлять пустые батчи
	}

	// Создаем JSON тело запроса
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics to JSON: %w", err)
	}

	// Сжимаем данные с помощью gzip
	var compressedData bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedData)
	if _, err := gzWriter.Write(jsonData); err != nil {
		gzWriter.Close()
		return fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return retrySend(func() error {
		url := fmt.Sprintf("%s/updates", c.baseURL)

		var body io.Reader
		var contentType string

		// Если задан крипто ключ, шифруем данные
		if c.cryptoKey != nil {
			encryptedData, err := c.cryptoKey.Encrypt(compressedData.Bytes())
			if err != nil {
				return fmt.Errorf("failed to encrypt data: %w", err)
			}
			body = bytes.NewReader(encryptedData)
			contentType = "application/json"
		} else {
			body = bytes.NewReader(compressedData.Bytes())
			contentType = "application/json"
		}

		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		// Добавляем заголовок X-Encrypted если данные зашифрованы
		if c.cryptoKey != nil {
			req.Header.Set("X-Encrypted", "true")
		}

		// Добавляем хеш, если ключ задан (до шифрования)
		if c.key != "" && c.cryptoKey == nil {
			hash := signature.CalculateHash(jsonData, c.key)
			req.Header.Set("HashSHA256", hash)
		}

		c.addXRealIPHeader(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
		}

		log.Printf("Successfully sent %d metrics batch to server", len(metrics))
		return nil
	})
}
