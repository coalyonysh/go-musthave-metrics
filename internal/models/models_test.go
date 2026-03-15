package models

import (
	"encoding/json"
	"testing"
)

func TestMetricJSONMarshal(t *testing.T) {
	delta := int64(100)

	m := Metric{
		ID:    "test_metric",
		MType: Counter,
		Delta: &delta,
		Value: nil,
		Hash:  "abc123",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Metric
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != m.ID {
		t.Errorf("expected ID %s, got %s", m.ID, unmarshaled.ID)
	}
	if unmarshaled.MType != m.MType {
		t.Errorf("expected MType %s, got %s", m.MType, unmarshaled.MType)
	}
	if unmarshaled.Delta == nil || *unmarshaled.Delta != *m.Delta {
		t.Errorf("expected Delta %d, got %v", *m.Delta, unmarshaled.Delta)
	}
	if unmarshaled.Hash != m.Hash {
		t.Errorf("expected Hash %s, got %s", m.Hash, unmarshaled.Hash)
	}
}

func TestMetricGaugeJSONMarshal(t *testing.T) {
	value := float64(98.5)

	m := Metric{
		ID:    "gauge_metric",
		MType: Gauge,
		Delta: nil,
		Value: &value,
		Hash:  "hash456",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Metric
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Value == nil || *unmarshaled.Value != *m.Value {
		t.Errorf("expected Value %f, got %v", *m.Value, unmarshaled.Value)
	}
}

func TestAuditEventJSONMarshal(t *testing.T) {
	event := AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1", "metric2", "metric3"},
		IPAddress: "192.168.1.1",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled AuditEvent
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Timestamp != event.Timestamp {
		t.Errorf("expected Timestamp %d, got %d", event.Timestamp, unmarshaled.Timestamp)
	}
	if len(unmarshaled.Metrics) != len(event.Metrics) {
		t.Errorf("expected %d metrics, got %d", len(event.Metrics), len(unmarshaled.Metrics))
	}
	if unmarshaled.IPAddress != event.IPAddress {
		t.Errorf("expected IPAddress %s, got %s", event.IPAddress, unmarshaled.IPAddress)
	}
}

func TestConstants(t *testing.T) {
	if Counter != "counter" {
		t.Errorf("expected Counter 'counter', got '%s'", Counter)
	}
	if Gauge != "gauge" {
		t.Errorf("expected Gauge 'gauge', got '%s'", Gauge)
	}
}

func TestMetricWithNilValues(t *testing.T) {
	m := Metric{
		ID:    "nil_test",
		MType: Gauge,
		Delta: nil,
		Value: nil,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Delta and Value should be omitted due to omitempty
	expected := `{"id":"nil_test","type":"gauge"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}
