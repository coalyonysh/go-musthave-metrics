package storage

import (
	"sync"
)

// SyncStorage обертка для Storage с поддержкой синхронного сохранения
type SyncStorage struct {
	Storage Storage
	saveFunc func()
	mu       sync.RWMutex
}

// NewSyncStorage создает новую обертку для Storage
func NewSyncStorage(s Storage, saveFunc func()) *SyncStorage {
	return &SyncStorage{
		Storage:  s,
		saveFunc: saveFunc,
	}
}

func (s *SyncStorage) SetGauge(name string, value float64) {
	s.mu.Lock()
	s.Storage.SetGauge(name, value)
	s.mu.Unlock()
	
	// Сохраняем синхронно, если функция сохранения задана
	if s.saveFunc != nil {
		s.saveFunc()
	}
}

func (s *SyncStorage) SetCounter(name string, value int64) {
	s.mu.Lock()
	s.Storage.SetCounter(name, value)
	s.mu.Unlock()
	
	// Сохраняем синхронно, если функция сохранения задана
	if s.saveFunc != nil {
		s.saveFunc()
	}
}

func (s *SyncStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Storage.GetGauge(name)
}

func (s *SyncStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Storage.GetCounter(name)
}

func (s *SyncStorage) GetAllGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Storage.GetAllGauges()
}

func (s *SyncStorage) GetAllCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Storage.GetAllCounters()
}

