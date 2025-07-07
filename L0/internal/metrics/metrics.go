package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status_code"})
	HttpRequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	})

	KafkaMessagesReceivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_received_total",
		Help: "Total number of Kafka messages received.",
	})
	KafkaMessagesInvalidTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_invalid_total",
		Help: "Total number of Kafka messages invalid.",
	})

	OrdersProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_processed_total",
		Help: "Total number of orders processed.",
	})

	CacheHitTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_hit_total",
		Help: "Total number of cache hits.",
	})
	CacheMissTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_miss_total",
		Help: "Total number of cache miss.",
	})
)
