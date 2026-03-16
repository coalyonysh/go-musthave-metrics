package agent

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/coalyonysh/go-musthave-metrics/api/proto"
	"github.com/coalyonysh/go-musthave-metrics/internal/models"
)

// MetricGRPCClient представляет gRPC клиент для отправки метрик
type MetricGRPCClient struct {
	client   proto.MetricsClient
	conn     *grpc.ClientConn
	clientIP string
}

// NewMetricGRPCClient создает новый gRPC клиент для метрик
func NewMetricGRPCClient(address string, clientIP string) (*MetricGRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &MetricGRPCClient{
		client:   proto.NewMetricsClient(conn),
		conn:     conn,
		clientIP: clientIP,
	}, nil
}

// SendMetricsBatch отправляет батч метрик через gRPC
func (c *MetricGRPCClient) SendMetricsBatch(metrics []models.Metric) error {
	protoMetrics := make([]*proto.Metric, 0, len(metrics))
	for _, m := range metrics {
		protoM := &proto.Metric{
			Id: m.ID,
		}

		if m.MType == models.Counter {
			protoM.Type = proto.Metric_COUNTER
			if m.Delta != nil {
				protoM.Delta = *m.Delta
			}
		} else if m.MType == models.Gauge {
			protoM.Type = proto.Metric_GAUGE
			if m.Value != nil {
				protoM.Value = *m.Value
			}
		}

		protoMetrics = append(protoMetrics, protoM)
	}

	req := &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	}

	// Создаем контекст с метаданными для передачи IP адреса
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Добавляем IP в метаданные
	md := metadata.Pairs("x-real-ip", c.clientIP)
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.client.UpdateMetrics(ctx, req)
	return err
}

// Close закрывает соединение
func (c *MetricGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getLocalIP возвращает локальный IP адрес
func getLocalIP() string {
	// Пытаемся получить IP через подключение к внешнему адресу
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
