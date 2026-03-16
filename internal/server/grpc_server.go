package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/coalyonysh/go-musthave-metrics/api/proto"
	"github.com/coalyonysh/go-musthave-metrics/internal/models"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
)

// MetricsServerImpl реализует интерфейс MetricsServer
type MetricsServerImpl struct {
	proto.UnimplementedMetricsServer
	storage storage.Storage
}

// NewMetricsServerImpl создает новый экземпляр MetricsServerImpl
func NewMetricsServerImpl(storage storage.Storage) *MetricsServerImpl {
	return &MetricsServerImpl{
		storage: storage,
	}
}

// UpdateMetrics обрабатывает запрос на обновление метрик
func (s *MetricsServerImpl) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if req == nil || req.Metrics == nil {
		return &proto.UpdateMetricsResponse{}, nil
	}

	metrics := make([]models.Metric, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		var mType string

		switch m.Type {
		case proto.Metric_COUNTER:
			mType = models.Counter
		case proto.Metric_GAUGE:
			mType = models.Gauge
		default:
			continue
		}

		metric := models.Metric{
			ID:    m.Id,
			MType: mType,
		}

		if mType == models.Counter {
			metric.Delta = &m.Delta
		} else if mType == models.Gauge {
			metric.Value = &m.Value
		}

		metrics = append(metrics, metric)
	}

	if len(metrics) > 0 {
		s.storage.SetMetricsBatch(metrics)
	}

	return &proto.UpdateMetricsResponse{}, nil
}

// trustedSubnetChecker проверяет, принадлежит ли IP к доверенной подсети
type trustedSubnetChecker struct {
	ipNet *net.IPNet
}

// newTrustedSubnetChecker создает новый checker с заданной подсетью
func newTrustedSubnetChecker(subnet string) (*trustedSubnetChecker, error) {
	if subnet == "" {
		return &trustedSubnetChecker{ipNet: nil}, nil
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	return &trustedSubnetChecker{ipNet: ipNet}, nil
}

// check проверяет IP адрес
func (c *trustedSubnetChecker) check(ip string) bool {
	if c.ipNet == nil {
		return true // если подсеть не задана, пропускаем все
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return c.ipNet.Contains(parsedIP)
}

// UnaryTrustedSubnetInterceptor создает interceptor для проверки подсети
func UnaryTrustedSubnetInterceptor(subnet string) (grpc.UnaryServerInterceptor, error) {
	checker, err := newTrustedSubnetChecker(subnet)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Получаем IP из метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		// Ищем x-real-ip в метаданных (grpc metadata keys are lowercase)
		values := md["x-real-ip"]
		if len(values) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip header")
		}

		clientIP := values[0]

		// Проверяем IP
		if !checker.check(clientIP) {
			return nil, status.Error(codes.PermissionDenied, "IP not in trusted subnet")
		}

		return handler(ctx, req)
	}, nil
}
