package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// Observer интерфейс для наблюдателей аудита
type Observer interface {
	Notify(event models.AuditEvent) error
}

// FileObserver реализует Observer для записи в файл
type FileObserver struct {
	filePath string
}

func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{filePath: filePath}
}

func (fo *FileObserver) Notify(event models.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	data = append(data, '\n')

	file, err := os.OpenFile(fo.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}

// URLOobserver реализует Observer для отправки по HTTP
type URLOobserver struct {
	url string
}

func NewURLOobserver(url string) *URLOobserver {
	return &URLOobserver{url: url}
}

func (uo *URLOobserver) Notify(event models.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	resp, err := http.Post(uo.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send audit event to URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit URL returned status: %d", resp.StatusCode)
	}

	return nil
}

// AuditService управляет наблюдателями и уведомляет их о событиях
type AuditService struct {
	observers []Observer
}

func NewAuditService() *AuditService {
	return &AuditService{observers: []Observer{}}
}

func (as *AuditService) AddObserver(observer Observer) {
	as.observers = append(as.observers, observer)
}

func (as *AuditService) Log(metrics []string, ipAddress string) {
	if len(as.observers) == 0 {
		return
	}

	event := models.AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddress,
	}

	for _, observer := range as.observers {
		if err := observer.Notify(event); err != nil {
		}
	}
}
