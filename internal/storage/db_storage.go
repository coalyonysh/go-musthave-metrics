package storage

import (
	"database/sql"
	"fmt"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (s *DBStorage) SetGauge(name string, value float64) {
	_, err := s.db.Exec(`
		INSERT INTO metrics (name, type, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (name, type) DO UPDATE SET value = EXCLUDED.value`,
		name, "gauge", value)
	if err != nil {
		fmt.Printf("Error setting gauge: %v\n", err)
	}
}

func (s *DBStorage) SetCounter(name string, value int64) {
	_, err := s.db.Exec(`
		INSERT INTO metrics (name, type, delta)
		VALUES ($1, $2, $3)
		ON CONFLICT (name, type) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta`,
		name, "counter", value)
	if err != nil {
		fmt.Printf("Error setting counter: %v\n", err)
	}
}

func (s *DBStorage) GetGauge(name string) (float64, bool) {
	var value float64
	err := s.db.QueryRow("SELECT value FROM metrics WHERE name = $1 AND type = $2", name, "gauge").Scan(&value)
	if err != nil {
		return 0, false
	}
	return value, true
}

func (s *DBStorage) GetCounter(name string) (int64, bool) {
	var value int64
	err := s.db.QueryRow("SELECT delta FROM metrics WHERE name = $1 AND type = $2", name, "counter").Scan(&value)
	if err != nil {
		return 0, false
	}
	return value, true
}

func (s *DBStorage) GetAllGauges() map[string]float64 {
	rows, err := s.db.Query("SELECT name, value FROM metrics WHERE type = $1", "gauge")
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err == nil {
			result[name] = value
		}
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return result
}

func (s *DBStorage) GetAllCounters() map[string]int64 {
	rows, err := s.db.Query("SELECT name, delta FROM metrics WHERE type = $1", "counter")
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err == nil {
			result[name] = value
		}
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return result
}
