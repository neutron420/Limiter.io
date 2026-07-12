package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"path", "method", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	RateLimiterDecisions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limiter_decisions_total",
			Help: "Total number of rate limiting decisions made by the engine",
		},
		[]string{"project_id", "decision"}, // decision = allowed / blocked
	)

	RedisLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_latency_seconds",
			Help:    "Latency of Redis operations in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.002, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{"operation"},
	)

	PostgresLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "postgres_latency_seconds",
			Help:    "Latency of PostgreSQL operations in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0},
		},
		[]string{"operation"},
	)

	KafkaPublishLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "kafka_publish_latency_seconds",
			Help:    "Latency of publishing events to Kafka in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		},
	)

	ErrorRate = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of application errors",
		},
		[]string{"type"},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_count",
			Help: "Current count of active registered users",
		},
	)

	ActiveProjects = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_projects_count",
			Help: "Current count of active projects",
		},
	)

	ActiveAPIKeys = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_apikeys_count",
			Help: "Current count of active API keys",
		},
	)
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconvItoa(c.Writer.Status())
		method := c.Request.Method

		// Ignore metrics path itself to avoid noise
		if path != "/metrics" && path != "/health" {
			HttpRequestsTotal.WithLabelValues(path, method, status).Inc()
			HttpRequestDuration.WithLabelValues(path, method, status).Observe(duration)
		}
	}
}

// Helper to avoid importing strconv just for Itoa
func strconvItoa(val int) string {
	switch val {
	case 200:
		return "200"
	case 201:
		return "201"
	case 204:
		return "204"
	case 400:
		return "400"
	case 401:
		return "401"
	case 403:
		return "403"
	case 404:
		return "404"
	case 429:
		return "429"
	case 500:
		return "500"
	default:
		// simple fallback (not high frequency)
		buf := make([]byte, 0, 10)
		for val > 0 {
			buf = append(buf, byte('0'+val%10))
			val /= 10
		}
		// reverse
		for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
			buf[i], buf[j] = buf[j], buf[i]
		}
		if len(buf) == 0 {
			return "0"
		}
		return string(buf)
	}
}
