package storage

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

func TestNewDBStorage(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)
	if storage == nil {
		t.Fatal("NewDBStorage returned nil")
	}
}

func TestDBStorage_SetGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs("test_gauge", "gauge", 1.5).
		WillReturnResult(sqlmock.NewResult(1, 1))

	storage.SetGauge("test_gauge", 1.5)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %v", err)
	}
}

func TestDBStorage_SetGauge_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs("test_gauge", "gauge", 1.5).
		WillReturnError(errors.New("database error"))

	// Should not panic, just log error
	storage.SetGauge("test_gauge", 1.5)
}

func TestDBStorage_SetCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs("test_counter", "counter", int64(10)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	storage.SetCounter("test_counter", 10)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %v", err)
	}
}

func TestDBStorage_GetGauge_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	rows := sqlmock.NewRows([]string{"value"}).
		AddRow(1.5)
	mock.ExpectQuery("SELECT value FROM metrics").
		WithArgs("test_gauge", "gauge").
		WillReturnRows(rows)

	value, found := storage.GetGauge("test_gauge")

	if !found {
		t.Error("expected to find gauge")
	}
	if value != 1.5 {
		t.Errorf("expected value 1.5, got %f", value)
	}
}

func TestDBStorage_GetGauge_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectQuery("SELECT value FROM metrics").
		WithArgs("nonexistent", "gauge").
		WillReturnError(sql.ErrNoRows)

	value, found := storage.GetGauge("nonexistent")

	if found {
		t.Error("expected not to find gauge")
	}
	if value != 0.0 {
		t.Errorf("expected value 0.0, got %f", value)
	}
}

func TestDBStorage_GetCounter_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	rows := sqlmock.NewRows([]string{"delta"}).
		AddRow(int64(100))
	mock.ExpectQuery("SELECT delta FROM metrics").
		WithArgs("test_counter", "counter").
		WillReturnRows(rows)

	value, found := storage.GetCounter("test_counter")

	if !found {
		t.Error("expected to find counter")
	}
	if value != 100 {
		t.Errorf("expected value 100, got %d", value)
	}
}

func TestDBStorage_GetCounter_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectQuery("SELECT delta FROM metrics").
		WithArgs("nonexistent", "counter").
		WillReturnError(sql.ErrNoRows)

	value, found := storage.GetCounter("nonexistent")

	if found {
		t.Error("expected not to find counter")
	}
	if value != 0 {
		t.Errorf("expected value 0, got %d", value)
	}
}

func TestDBStorage_GetAllGauges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("gauge1", 1.0).
		AddRow("gauge2", 2.5)
	mock.ExpectQuery("SELECT name, value FROM metrics").
		WithArgs("gauge").
		WillReturnRows(rows)

	result := storage.GetAllGauges()

	if len(result) != 2 {
		t.Errorf("expected 2 gauges, got %d", len(result))
	}
	if result["gauge1"] != 1.0 {
		t.Errorf("expected gauge1 = 1.0, got %f", result["gauge1"])
	}
	if result["gauge2"] != 2.5 {
		t.Errorf("expected gauge2 = 2.5, got %f", result["gauge2"])
	}
}

func TestDBStorage_GetAllGauges_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectQuery("SELECT name, value FROM metrics").
		WillReturnError(errors.New("database error"))

	result := storage.GetAllGauges()

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestDBStorage_GetAllCounters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	rows := sqlmock.NewRows([]string{"name", "delta"}).
		AddRow("counter1", int64(10)).
		AddRow("counter2", int64(20))
	mock.ExpectQuery("SELECT name, delta FROM metrics").
		WithArgs("counter").
		WillReturnRows(rows)

	result := storage.GetAllCounters()

	if len(result) != 2 {
		t.Errorf("expected 2 counters, got %d", len(result))
	}
	if result["counter1"] != 10 {
		t.Errorf("expected counter1 = 10, got %d", result["counter1"])
	}
	if result["counter2"] != 20 {
		t.Errorf("expected counter2 = 20, got %d", result["counter2"])
	}
}

func TestDBStorage_GetAllCounters_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectQuery("SELECT name, delta FROM metrics").
		WillReturnError(errors.New("database error"))

	result := storage.GetAllCounters()

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestDBStorage_SetMetricsBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO metrics").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO metrics").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	metrics := []models.Metric{
		{ID: "gauge1", MType: "gauge", Value: func() *float64 { v := 1.0; return &v }()},
		{ID: "counter1", MType: "counter", Delta: func() *int64 { v := int64(10); return &v }()},
	}

	err = storage.SetMetricsBatch(metrics)
	if err != nil {
		t.Errorf("SetMetricsBatch returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %v", err)
	}
}

func TestDBStorage_SetMetricsBatch_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	// For empty slice, Begin is called but nothing else
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = storage.SetMetricsBatch([]models.Metric{})
	if err != nil {
		t.Errorf("SetMetricsBatch returned error for empty slice: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations not met: %v", err)
	}
}

func TestIsRetriableDBError_NetworkError(t *testing.T) {
	err := errors.New("connection refused")
	result := isRetriableDBError(err)
	if result {
		t.Error("expected false for generic error")
	}
}

func TestDBStorage_SetMetricsBatch_NonRetriableError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	storage := NewDBStorage(db)

	// Non-retriable error should return immediately
	mock.ExpectBegin().
		WillReturnError(errors.New("some error"))

	metrics := []models.Metric{
		{ID: "gauge1", MType: "gauge", Value: func() *float64 { v := 1.0; return &v }()},
	}

	err = storage.SetMetricsBatch(metrics)
	if err == nil {
		t.Error("expected error for non-retriable error")
	}
}
